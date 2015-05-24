package goto_rpc

import "net"
import "time"
import proto "encoding/protobuf/proto"

type Client struct {
	RpcConn
}

func NewClient(c net.Conn) *Client {
	method_map := make(MethodMap)
	rc := &Client{
		*NewRpcConn(c, 3 * time.Second, 3 * time.Second, &method_map)}
	rc.active()
	return rc
}

func (this *Client) SetTimeout(send_timeout, recv_timeout time.Duration) {
	this.send_timeout = send_timeout
	this.recv_timeout = recv_timeout
}

func (this *Client) Call(method string, request proto.Message) (response proto.Message, err error) {

	c_rsp := make(chan proto.Message)
	c_err := make(chan error)
	this.AsynCall(method, request, func(err error, rsp proto.Message) {
		if err != nil {
			c_err <- err
		} else {
			c_rsp <- rsp
		}
    })
	select {
	case response = <-c_rsp:
		return 
	case err = <-c_err:
		return 
	}
}
