package transport

import "net"
import "time"
import proto "encoding/protobuf/proto"

type RpcClient struct {
	RpcConn
}

func NewRpcClient(c net.Conn) *RpcClient {
	method_map := make(MethodMap)
	rc := &RpcClient{
		*NewRpcConn(c, 3 * time.Second, 3 * time.Second, &method_map)}
	rc.active()
	return rc
}

func (this *RpcClient) SetTimeout(send_timeout, recv_timeout time.Duration) {
	this.send_timeout = send_timeout
	this.recv_timeout = recv_timeout
}

func (this *RpcClient) Call(method string, request proto.Message) (response proto.Message, err error) {

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
