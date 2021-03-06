package goto_rpc

import (
	"net"
	"strconv"
)
import "time"
import "runtime"

import "sync"
import proto "encoding/protobuf/proto"

type send_chan chan IPackage

type IRpcConn interface {
	reply(pkg IPackage, rsp_status byte, response proto.Message) error
	go_reply(pkg IPackage, rsp_status byte, response proto.Message) error
}

type RpcConn struct {
	low_layer     net.Conn
	send_queue    send_chan
	channel_ok    bool
	send_timeout  time.Duration
	recv_timeout  time.Duration
	recv_buf      []byte
	chan_lock     sync.RWMutex
	ctx_map       map[uint32]ICallContext
	start_seq_num uint32
	method_map    *MethodMap
}

func NewRpcConn(low_layer net.Conn,
	send_timeout time.Duration, recv_timeout time.Duration,
	method_map *MethodMap, send_buf_size int) *RpcConn {

	return &RpcConn{low_layer,
		make(send_chan, send_buf_size),
		true,
		send_timeout,
		recv_timeout,
		make([]byte, 4096),
		sync.RWMutex{},
		make(map[uint32]ICallContext),
		0,
		method_map}
}

func (c *RpcConn) active() {
	go c.do_write()
	go c.do_read()
}

func (c *RpcConn) do_write() {
	for {
		ctx, ok := <-c.send_queue
		if ok == false {
			break
		}

		is_inserted_ctx_map := false
		if ctx.GetRpcType() == RpcType_Request {
			if call_ctx, ok := ctx.(ICallContext); ok {
				c.start_seq_num++
				ctx.SetSeqNum(c.start_seq_num)

				old_ctx, ok := c.ctx_map[ctx.GetSeqNum()]
				if ok == true {
					old_ctx.CallError(NewError(RpcError_Overwrite))
				}

				c.ctx_map[ctx.GetSeqNum()] = call_ctx

				if c.recv_timeout > 0 {
					time.AfterFunc(c.recv_timeout, func() {
						if ctx.CallError(NewError(RpcError_RecvTimeout)) == true {
							delete(c.ctx_map, ctx.GetSeqNum())
						}
					})
				}

				is_inserted_ctx_map = true
			}
		}

		b, e := ctx.Marshal()
		if e != nil {
			if is_inserted_ctx_map == true {
				delete(c.ctx_map, ctx.GetSeqNum())
			}
			ctx.CallError(e)
			continue
		}

		_, e = c.low_layer.Write(b)
		if e != nil {
			if is_inserted_ctx_map == true {
				delete(c.ctx_map, ctx.GetSeqNum())
			}
			ctx.CallError(e)
			break
		}
	}

	c.Shutdown()
	for {
		ctx, ok := <-c.send_queue
		if ok == false {
			break
		}

		ctx.CallError(NewError(RpcError_NotEstab))
	}
}

func (c *RpcConn) do_read() {
	pos := 0
	for {
		n, e := c.low_layer.Read(c.recv_buf[pos:])
		if e != nil {
			break
		}
		pos += n
		nn := 0
	Retry:
		pkg, body, consume, err := Unmarshal_PackageHead(c.recv_buf[nn:pos])
		if err != nil {
			// data parse error
			break
		}
		if pkg != nil {
			nn += consume
			c.do_package(pkg, body)
			goto Retry
		}
		copy(c.recv_buf, c.recv_buf[nn:pos])
		pos -= nn
		runtime.Gosched()
	}

	for _, ctx := range c.ctx_map {
		ctx.CallError(NewError(RpcError_NotEstab))
	}

	logger.Printf("Disconnect %s", c.low_layer.RemoteAddr().String())
	c.close_read()
}

func (c *RpcConn) do_package(pkg *Package, body []byte) {
	logger.Printf("recv package: {type=%s, status=%d, seq_num=%d, method=%s, has_body=%s}",
		RpcTypeToString(pkg.rpc_type), pkg.rsp_status, pkg.seq_num, pkg.method,
		strconv.FormatBool(pkg.body == nil))

	if pkg.rpc_type == RpcType_Response {
		ctx, ok := c.ctx_map[pkg.seq_num]
		if ok == false {
			// request was timeout, discard it.
			logger.Fatalf("not in ctx map. size=%d", len(c.ctx_map))
			return
		}

		method_info, ok := (*c.method_map)[pkg.method]
		if ok == false {
			// error response.
			logger.Fatalf("hasn't method info. method_map size=%d", len(*c.method_map))
			return
		}

		pkg.body = method_info.rsp_factory()
		if pkg.body == nil {
			// create response struct error.
			logger.Fatalf("create response struct error.")
			return
		}

		e := Unmarshal_Body(pkg, body)
		if e != nil {
			logger.Fatalf("parse response body error.")
			return
		}

		delete(c.ctx_map, pkg.seq_num)
		ctx.Call(nil, pkg.body)
	} else {
		method_info, ok := (*c.method_map)[pkg.method]
		if ok == false || method_info.service_func == nil {
			// unkown method
			if pkg.rpc_type == RpcType_Request {
				c.reply(pkg, RpcError_NoMethod, nil)
            }

			return
		}

		pkg.body = method_info.req_factory()
		if pkg.body == nil {
			// create request struct error.
			logger.Fatalf("create request struct error.")
			return
		}

		e := Unmarshal_Body(pkg, body)
		if e != nil {
			logger.Fatalf("parse request body error.")
			return
		}

		ctx := NewContext(pkg, c)
		method_info.service_func(ctx, pkg.body)
	}
}

func (c *RpcConn) AsynCall(method string, request proto.Message, cb RpcCallback) (e error) {
	if c.channel_ok == false {
		return NewError(RpcError_NotEstab)
	}

	req := NewCallContext(method, request, cb)

	defer func(e *error) {
		if err := recover(); err != nil {
			*e = NewError(RpcError_NotEstab)
        }
	}(&e)

	select {
	case c.send_queue <- req:
		if c.send_timeout > 0 {
			time.AfterFunc(c.send_timeout, func() {
				req.CallError(NewError(RpcError_SendTimeout))
			})
        }
		return 
	default:
		return NewError(RpcError_BufferFull)
	}
}

func (c *RpcConn) GoAsynCall(method string, request proto.Message, cb RpcCallback) (e error) {
	if c.channel_ok == false {
		return NewError(RpcError_NotEstab)
	}

	req := NewCallContext(method, request, cb)

	//try send
	defer func(e *error) {
		if err := recover(); err != nil {
			*e = NewError(RpcError_NotEstab)
        }
	}(&e)

	select {
	case c.send_queue <- req:
		if c.send_timeout > 0 {
			time.AfterFunc(c.send_timeout, func() {
				req.CallError(NewError(RpcError_SendTimeout))
			})
        }
		return 
	default:
	}

	go func() {
		defer func(e *error) {
			if err := recover(); err != nil {
				*e = NewError(RpcError_NotEstab)
			}
		}(&e)

		if c.send_timeout > 0 {
			t := time.After(c.send_timeout)
			select {
			case c.send_queue <- req:
				time.AfterFunc(c.send_timeout, func() {
					req.CallError(NewError(RpcError_SendTimeout))
				})
				return 
			case <-t:
				cb(NewError(RpcError_SendTimeout), nil)
			}
        } else {
			c.send_queue <- req
        }
	}()

	return 
}

func (c *RpcConn) reply(pkg IPackage, rsp_status byte, response proto.Message) (e error) {
	if pkg.GetRpcType() != RpcType_Request {
		return
	}

	if c.channel_ok == false {
		return NewError(RpcError_NotEstab)
	}

	rsp := NewResponsePackage(rsp_status, pkg.GetSeqNum(), pkg.GetMethod(), response)

	defer func(e *error) {
		if err := recover(); err != nil {
			*e = NewError(RpcError_SendTimeout)
        }
	}(&e)

	select {
	case c.send_queue <- rsp:
		return
	default:
		return NewError(RpcError_BufferFull)
	}
}


func (c *RpcConn) go_reply(pkg IPackage, rsp_status byte, response proto.Message) (e error) {
	if pkg.GetRpcType() != RpcType_Request {
		return
	}

	if c.channel_ok == false {
		return NewError(RpcError_NotEstab)
	}

	rsp := NewResponsePackage(rsp_status, pkg.GetSeqNum(), pkg.GetMethod(), response)

	// try reply
	defer func(e *error) {
		if err := recover(); err != nil {
			*e = NewError(RpcError_SendTimeout)
        }
	}(&e)

	select {
	case c.send_queue <- rsp:
		return
	default:
	}

	go func() {
		defer func(e *error) {
			if err := recover(); err != nil {
				*e = NewError(RpcError_SendTimeout)
			}
		}(&e)

		if c.send_timeout > 0 {
			t := time.After(c.send_timeout)
			select {
			case c.send_queue <- rsp:
			case <-t:
			}
        } else {
			c.send_queue <- rsp
        }
	}()

	return 
}

func (c *RpcConn) close_write() {
	tcp_conn, ok := c.low_layer.(*net.TCPConn)
	if ok {
		tcp_conn.CloseWrite()
	} else {
		c.low_layer.Close()
	}
}

func (c *RpcConn) close_read() {
	tcp_conn, ok := c.low_layer.(*net.TCPConn)
	if ok {
		tcp_conn.CloseRead()
	} else {
		c.low_layer.Close()
	}
}

func (c *RpcConn) Shutdown() {
	if c.channel_ok {
		c.channel_ok = false

		c.chan_lock.Lock()
		defer c.chan_lock.Unlock()

		if c.channel_ok {
			close(c.send_queue)
			c.close_write()
        }
	}
}

func (c *RpcConn) Close() {
	c.low_layer.Close()
}

