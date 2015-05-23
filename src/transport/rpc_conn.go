package transport

import "net"
import "time"
import "common"
import "sync"

type request_chan chan *RpcRequest

type RpcConn struct {
	low_layer     net.Conn
	request_queue request_chan
	channel_ok    bool
	send_timeout  time.Duration
	recv_timeout  time.Duration
	recv_buf      []byte
	chan_lock     sync.RWMutex
	request_map	  map[uint32]*RpcRequest
	seq_num	      uint32
}

func NewRpcConn(low_layer net.Conn,
	send_timeout time.Duration,
	recv_timeout time.Duration,
	request_queue_c int) *RpcConn {

	return &RpcConn{low_layer,
		make(request_chan, request_queue_c),
		true,
		send_timeout,
		recv_timeout,
		make([]byte, 0),
		sync.RWMutex{},
		make(map[uint32]*RpcRequest),
		0,
	}
}

func (c *RpcConn) Active() {
	// write
	go func() {
		for {
			ctx, ok := <-c.request_queue
			if ok == false {
				break
			}

			b, err := ctx.ToBytes()
			if err != nil {
				go ctx.Call(err)
				continue
			}

			_, e := c.low_layer.Write(b)
			if e != nil {
				go ctx.Call(err)
				break
			}

			c.seq_num ++
			ctx.seq_num = c.seq_num
			old_ctx, ok := c.request_map[ctx.seq_num]
			if ok == true {
				go old_ctx.Call(common.NewError(common.RpcError_Overwrite, ""))
            }

			c.request_map[ctx.seq_num] = ctx
			time.AfterFunc(c.recv_timeout, func(){
				if ctx.Call(common.NewError(common.RpcError_RecvTimeout, "")) == true {
					delete(c.request_map, ctx.seq_num)
				}
			})
		}

		go c.Shutdown()
		for {
			ctx, ok := <-c.request_queue
			if ok == false {
				break
			}

			go ctx.Call(common.NewError(common.RpcError_NotConn, ""))
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
			head, body, length, err := ParsePackage(c.recv_buf)
			if err != nil {
				// data parse error
				break
            }

			if head != nil {
				c.recv_buf = c.recv_buf[length:]
				go c.DoPackage(head, body)
				goto Retry
            }
        }

		for _, ctx := range c.request_map {
			go ctx.Call(common.NewError(common.RpcError_NotConn, ""))
        }

		c.low_layer.Close()
    }()
}

func (c *RpcConn) DoPackage(head *RpcPackageHead, body []byte) {

}

func (c *RpcConn) AsynCall(method string, body IRpcData, cb RpcCallback) error {
	if c.channel_ok == false {
		return common.NewError(common.RpcError_NotConn, "")
	}

	request := NewRpcRequest(method, body, cb)
	t := time.After(c.send_timeout)

	c.chan_lock.RLock()
	defer c.chan_lock.RUnlock()
	select {
	case c.request_queue <- request:
		return nil
	case <-t:
		return common.NewError(common.RpcError_SendTimeout, "")
	}
}

func (c *RpcConn) Shutdown() {
	if c.channel_ok == true {
		c.channel_ok = false

		c.chan_lock.Lock()
		defer c.chan_lock.Unlock()
		close(c.request_queue)
	}
}

func (c *RpcConn) ForceClose() error {
	return c.low_layer.Close()
}
