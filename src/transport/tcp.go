package transport

import "net"

type buf []byte
type buf_channel chan buf

type TcpTransport interface {
	net.TcpConn
	Remote_addr net.TCPAddr
	send_list buf_channel
	recv_buf buf
	is_closing bool
}

func NewTcpTransport(addr net.TCPAddr, buf_length int) *TcpTransport {
	obj := &TcpTransport{addr, make(buf_channel, buf_length)}
	return obj
}

func (c *TcpTransport) Active() {
	go func() {
		for {
			b, ok <- c.send_list
			if ok == false {
				break
			}

		Retry:
			n, e := c.net.TcpConn.Write(b)
			if e != nil {
				break
			}

			if n < len(b) {
				b = b[n:]
				n = len(b)
				goto Retry
            }
        }
    }
}

func (c *TcpTransport) Write(b []byte) (error) {
	c.send_list <- b
}

