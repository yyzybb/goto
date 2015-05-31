package goto_rpc

import "net"
import "time"
import proto "encoding/protobuf/proto"

type IPolicy interface {
	// set notify callback function
	SetNotifyCb(up_conn, down_conn func(string))

	// connect to server with blocking
	GetConn(key string) net.Conn
}

type ClientGroup struct {
	clients			map[string]*Client
	connecting		map[string]int
	policy			IPolicy
	method_map		*MethodMap
	send_buf_size   int
	send_timeout	time.Duration
	recv_timeout	time.Duration
	robin_index		int
}

func (this *ClientGroup) up_conn(key string) {
	if _, ok := this.connecting[key]; ok {
		return 
    }

	c := this.policy.GetConn(key)
	if c == nil { return }

	this.clients[key] = NewClientByMethodMap(c, this.send_buf_size, this.method_map)
}

func (this *ClientGroup) down_conn(key string) {
	delete(this.clients, key)
	delete(this.connecting, key)
}

func (this *ClientGroup) get_conn() *Client {
	i := 0
	robin_index ++
	if robin_index >= len(this.clients) {
		robin_index = 0
	}

	for ; key, c := range this.clients; i++ {
		if i == robin_index {
			return c
        }
	}
}

func (this *ClientGroup) AsynCall(method string, request proto.Message, cb RpcCallback) (e error) {

}

