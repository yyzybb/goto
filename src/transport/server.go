package transport

import "net"
import "time"
import "common"

type RpcServer struct {
	listener		net.Listener
	method_map		MethodMap
	send_timeout	time.Duration
	recv_timeout	time.Duration
}

func NewRpcServer(listener net.Listener) *RpcServer {
	return &RpcServer{listener, make(MethodMap), 3 * time.Second, 3 * time.Second}
}

func (this *RpcServer) SetTimeout(send_timeout, recv_timeout time.Duration) {
	this.send_timeout = send_timeout
	this.recv_timeout = recv_timeout
}

func (this *RpcServer) AddServiceFunc(method string, fn RpcServiceFunc,
	req_factory RpcMessageFactoryFunc, rsp_factory RpcMessageFactoryFunc) (err error) {

	if _, exists := this.method_map[method]; exists {
		err = common.NewError(common.RpcError_RepeatMethod)
		return 
    }

	method_info := &MethodInfo{method, fn, req_factory, rsp_factory}
	this.method_map[method] = method_info
	return
}

func (this *RpcServer) Start() {
	for {
		conn, e := this.listener.Accept()
		if e != nil {
			continue
		}

		rpc_conn := NewRpcConn(conn, this.send_timeout, this.recv_timeout, &this.method_map)
		rpc_conn.active()
    }
}
