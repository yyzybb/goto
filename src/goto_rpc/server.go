package goto_rpc

import "net"
import "time"

type Server struct {
	listener		net.Listener
	method_map		MethodMap
	send_timeout	time.Duration
	recv_timeout	time.Duration
	send_buf_size   int
}

func NewServer(listener net.Listener, send_buf_size int) *Server {
	return &Server{listener, make(MethodMap), 3 * time.Second, 3 * time.Second, send_buf_size}
}

func (this *Server) SetTimeout(send_timeout, recv_timeout time.Duration) {
	this.send_timeout = send_timeout
	this.recv_timeout = recv_timeout
}

func (this *Server) AddServiceFunc(method string, fn RpcServiceFunc,
	req_factory RpcMessageFactoryFunc, rsp_factory RpcMessageFactoryFunc) (err error) {

	if _, exists := this.method_map[method]; exists {
		err = NewError(RpcError_RepeatMethod)
		return 
    }

	logger.Printf("AddServiceFunc [%s]", method)
	method_info := &MethodInfo{method, fn, req_factory, rsp_factory}
	this.method_map[method] = method_info
	return
}

func (this *Server) Start() {
	for {
		conn, e := this.listener.Accept()
		if e != nil {
			continue
		}

		logger.Printf("Accept %s", conn.RemoteAddr().String())
		rpc_conn := NewRpcConn(conn, this.send_timeout, this.recv_timeout, &this.method_map, this.send_buf_size)
		rpc_conn.active()
    }
}
