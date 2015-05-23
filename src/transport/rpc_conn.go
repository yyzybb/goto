package transport

import "net"
import "time"
import "errors"

type RpcPackage interface {
	Buf() *[]byte
	Done(err error)
}

type buf_channel chan RpcPackage

type RpcConn struct {
	*net.Conn
	send_list	buf_channel
	channel_ok	bool
	send_timeout	time.Duration
	recv_buf	[]byte
}

func NewRpcConn(low_layer *net.Conn, timeout time.Duration, max_package_size int, send_list_c int) *RpcConn {
	obj := &RpcConn{low_layer,
		make(buf_channel, send_list_c),
		true,
		timeout,
		make([]byte, max_package_size)}

	return obj
}

func (c *RpcConn) Active() {
	// write
	go func() {
		for {
			b, ok := <-c.send_list
			if ok == false {
				break
			}

			_, e := c.TCPConn.Write(b)
			if e != nil {
				break
			}
		}

		c.CloseWrite()
	}()
}

func (c *RpcConn) Loop() {
	// read
}

func (c *RpcConn) Write(b []byte) error {
	t := time.NewTimer(c.send_timeout)
	select {
	case c.send_list <- b:
		return nil
	case <-t.C:
		return errors.New("Timeout")
	}
}

func (c *RpcConn) Shutdown() {
	close(c.send_list)
}
