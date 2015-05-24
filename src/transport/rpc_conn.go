package transport

import "net"
import "time"
import "common"
import "sync"
import proto "encoding/protobuf/proto"

type send_chan chan IPackage

type IRpcConn interface {
	reply(pkg IPackage, rsp_status byte, response proto.Message) error
}

type RpcConn struct {
	low_layer     net.Conn
	send_queue	  send_chan
	channel_ok    bool
	send_timeout  time.Duration
	recv_timeout  time.Duration
	recv_buf      []byte
	chan_lock     sync.RWMutex
	ctx_map		  map[uint32]ICallContext
	start_seq_num uint32
	method_map	  *MethodMap
}

func NewRpcConn(low_layer net.Conn,
		send_timeout time.Duration,
		recv_timeout time.Duration,
		method_map	  *MethodMap) *RpcConn {

	return &RpcConn{low_layer,
		make(send_chan),
		true,
		send_timeout,
		recv_timeout,
		make([]byte, 0),
		sync.RWMutex{},
		make(map[uint32]ICallContext),
		0,
		method_map }
}

func (c *RpcConn) active() {
	// write
	go func() {
		for {
			ctx, ok := <-c.send_queue
			if ok == false {
				break
			}

			b, e := ctx.Marshal()
			if e != nil {
				ctx.CallError(e)
				continue
			}

			_, e = c.low_layer.Write(b)
			if e != nil {
				ctx.CallError(e)
				break
			}

			if ctx.GetRpcType() != RpcType_Request {
				continue
            }

			if call_ctx, ok := ctx.(ICallContext); ok {
				c.start_seq_num++
				ctx.SetSeqNum(c.start_seq_num)

				old_ctx, ok := c.ctx_map[ctx.GetSeqNum()]
				if ok == true {
					old_ctx.CallError(common.NewError(common.RpcError_Overwrite))
				}

				c.ctx_map[ctx.GetSeqNum()] = call_ctx
				time.AfterFunc(c.recv_timeout, func(){
					if ctx.CallError(common.NewError(common.RpcError_RecvTimeout)) == true {
						delete(c.ctx_map, ctx.GetSeqNum())
					}
				})
            }
		}

		go c.Shutdown()
		for {
			ctx, ok := <-c.send_queue
			if ok == false {
				break
			}

			ctx.CallError(common.NewError(common.RpcError_NotEstab))
		}

		c.low_layer.Close()
	}()

	// read
	go func() {
		for {
			b := make([]byte, 4096)
			n, e := c.low_layer.Read(b)
			if e != nil {
				break
            }

			c.recv_buf = append(c.recv_buf, b[:n]...)
		Retry:
			pkg, body, consume, err := Unmarshal_PackageHead(c.recv_buf)
			if err != nil {
				// data parse error
				break
            }

			if pkg != nil {
				c.recv_buf = c.recv_buf[consume:]
				c.do_package(pkg, body)
				goto Retry
            }
        }

		for _, ctx := range c.ctx_map {
			ctx.CallError(common.NewError(common.RpcError_NotEstab))
        }

		c.low_layer.Close()
    }()
}

func (c *RpcConn) do_package(pkg *Package, body []byte) {
	if pkg.rpc_type == RpcType_Response {
		ctx, ok := c.ctx_map[pkg.seq_num]
		if ok == false {
			// request was timeout, discard it.
			return 
		}

		method_info, ok := (*c.method_map)[pkg.method]
		if ok == false {
			// error response.
			return 
		}

		pkg.body = method_info.rsp_factory()
		if pkg.body == nil {
			// create response struct error.
			return 
		}

		e := Unmarshal_Body(pkg, body)
		if e != nil {
			return
		}

		ctx.Call(nil, pkg.body)
		delete(c.ctx_map, pkg.seq_num)
    } else {
		method_info, ok := (*c.method_map)[pkg.method]
		if ok == false {
			// unkown method
			if pkg.rpc_type == RpcType_Request {
				go c.reply(pkg, common.RpcError_NoMethod, nil)
            }

			return
		}

		pkg.body = method_info.req_factory()
		if pkg.body == nil {
			// create request struct error.
			return 
		}

		e := Unmarshal_Body(pkg, body)
		if e != nil {
			return
		}

		ctx := NewContext(pkg, c)
		go method_info.service_func(ctx, pkg.body)
    }
}

func (c *RpcConn) AsynCall(method string, request proto.Message, cb RpcCallback) error {
	if c.channel_ok == false {
		return common.NewError(common.RpcError_NotEstab)
	}

	req := NewCallContext(method, request, cb)
	t := time.After(c.send_timeout)

	c.chan_lock.RLock()
	defer c.chan_lock.RUnlock()
	if c.channel_ok == false {
		return common.NewError(common.RpcError_NotEstab)
	}

	select {
	case c.send_queue <- req:
		return nil
	case <-t:
		return common.NewError(common.RpcError_SendTimeout)
	}
}

func (c *RpcConn) reply(pkg IPackage, rsp_status byte, response proto.Message) error {
	if pkg.GetRpcType() != RpcType_Request {
		return nil
    }

	if c.channel_ok == false {
		return common.NewError(common.RpcError_NotEstab)
	}

	rsp := NewResponsePackage(rsp_status, pkg.GetSeqNum(), pkg.GetMethod(), response)
	t := time.After(c.send_timeout)

	c.chan_lock.RLock()
	defer c.chan_lock.RUnlock()
	if c.channel_ok == false {
		return common.NewError(common.RpcError_NotEstab)
	}

	select {
	case c.send_queue <- rsp:
		return nil
	case <-t:
		return common.NewError(common.RpcError_SendTimeout)
	}
}

func (c *RpcConn) Shutdown() {
	if c.channel_ok == true {
		c.channel_ok = false

		c.chan_lock.Lock()
		defer c.chan_lock.Unlock()
		close(c.send_queue)
	}
}

func (c *RpcConn) ForceClose() error {
	return c.low_layer.Close()
}
