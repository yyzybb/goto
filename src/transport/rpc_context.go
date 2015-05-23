package transport

import "common"

type IRpcData interface {
	ToBytes() []byte
}

type RpcPackageHead struct {
	need_response bool
	seq_num uint32
	method string
}
func (this *RpcPackageHead) Method() string {
	return this.method
}

func NewRpcPackageHead(b []byte) (head *RpcPackageHead, nbytes int) {
	head = &RpcPackageHead{}
	if b[3] == 0 {
		head.need_response = false
    } else {
		head.need_response = true
    }
	head.seq_num = uint32(b[4]) << 24 + uint32(b[5]) << 16 + uint32(b[6]) << 8 + uint32(b[7])
	head.method = string(b[9:b[8] + 9])
	nbytes = int(b[8]) + 8
	return 
}

type RpcPackage struct {
	RpcPackageHead
	body IRpcData
}
func NewRpcPackage(need_response bool, method string, body IRpcData) *RpcPackage {
	return &RpcPackage{RpcPackageHead{need_response, 0, method}, body}
}
func (this *RpcPackage) ToBytes() ([]byte, error) {
	b := make([]byte, 1 + 2 + 1 + 4 + 1 + len(this.method))
	b = append(b, 0xF8)
	b = append(b, byte(0), byte(0))
	if this.need_response == true {
		b = append(b, byte(1))
    } else {
		b = append(b, byte(0))
    }
	b = append(b, byte(this.seq_num >> 24))
	b = append(b, byte(this.seq_num >> 16 & 0xff))
	b = append(b, byte(this.seq_num >> 8 & 0xff))
	b = append(b, byte(this.seq_num & 0xff))
	b = append(b, byte(len(this.method)))
	b = append(b, this.method...)
	b = append(b, this.body.ToBytes()...)
	length := len(b)
	if length > 0xffff {
		return make([]byte, 0), common.NewError(common.RpcError_PackTooLarge, "")
    }
	b[1] = byte(length >> 8)
	b[2] = byte(length & 0xff)
	return b, nil
}
func ParsePackage(b []byte) (head *RpcPackageHead, body []byte, length int, err error) {
	if len(b) <= 5 {
		return 
    }

	if b[0] != 0xF8 {
		err = common.NewError(common.RpcError_Fatal, "")
		return 
    }

	length = int(b[1]) << 8 + int(b[2])
	if len(b) < length {
		n := 0
		head, n = NewRpcPackageHead(b)
		body = b[n:length]
		return 
	}

	return 
}

type RpcCallback func(error)
type RpcRequest struct {
	*RpcPackage

	is_valid bool
	cb RpcCallback
}
func NewRpcRequest(method string, body IRpcData, cb RpcCallback) *RpcRequest {
	rc := &RpcRequest{NewRpcPackage(cb != nil, method, body), true, cb}
	return rc
}
func (this *RpcRequest) Call(err error) bool {
	if this.is_valid == false {
		return false
    }

	this.is_valid = false
	this.cb(err)
	return true
}
