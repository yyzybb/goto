package goto_rpc

import "net"
import "time"
import proto "encoding/protobuf/proto"

type Client struct {
	RpcConn
	method_map		MethodMap
}

func NewClient(c net.Conn) *Client {
	rc := &Client{
		*NewRpcConn(c, 3 * time.Second, 3 * time.Second, nil), make(MethodMap)}
	rc.RpcConn.method_map = &rc.method_map
	rc.active()
	return rc
}

func (this *Client) SetTimeout(send_timeout, recv_timeout time.Duration) {
	this.send_timeout = send_timeout
	this.recv_timeout = recv_timeout
}

func (this *Client) AddServiceInfo(method string, req_factory RpcMessageFactoryFunc,
	rsp_factory RpcMessageFactoryFunc) (err error) {

	if _, exists := this.method_map[method]; exists {
		err = NewError(RpcError_RepeatMethod)
		return 
    }

	logger.Printf("AddServiceInfo [%s]", method)
	method_info := &MethodInfo{method, nil, req_factory, rsp_factory}
	this.method_map[method] = method_info
	return
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
