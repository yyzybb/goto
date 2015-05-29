package goto_rpc

import "common"
import "strconv"
import proto "encoding/protobuf/proto"

const (
	RpcType_Request byte = iota
	RpcType_Response
	RpcType_Oneway
)

func RpcTypeToString(rpc_type byte) string {
	switch rpc_type {
	case RpcType_Request: return "RpcType_Request"
	case RpcType_Response: return "RpcType_Response"
	case RpcType_Oneway: return "RpcType_Oneway"
	default: return strconv.Itoa(int(rpc_type))
    }
}

type IPackage interface {
	GetRpcType() byte
	GetMethod() string
	GetBody() proto.Message

	SetSeqNum(uint32)
	GetSeqNum() uint32
	Marshal() ([]byte, error)
	CallError(error) bool
}

type Package struct {
	rpc_type byte
	rsp_status byte
	seq_num uint32
	method string
	body proto.Message
}

type IContext interface {
	IPackage
	Reply(status byte, response proto.Message) error
}

// definition for IContext
type Context struct {
	Package
	conn IRpcConn
}

type ICallContext interface {
	IPackage
	Call(err error, response proto.Message) bool
}

type RpcCallback func(error, proto.Message)

// definition for CallContext
type CallContext struct {
	Package

	called bool
	cb RpcCallback
}

/// ------------------------- Package
func NewRequestPackage(rpc_type byte, method string, body proto.Message) *Package {
	return &Package{rpc_type, common.RpcError_Ok, 0, method, body}
}
func NewResponsePackage(rsp_status byte, seq_num uint32, method string, body proto.Message) *Package {
	return &Package{RpcType_Response, rsp_status, seq_num, method, body}
}
func (this *Package) GetRpcType() byte {
	return this.rpc_type
}
func (this *Package) SetSeqNum(seq_num uint32) {
	this.seq_num = seq_num
}

func (this *Package) GetSeqNum() uint32 {
	return this.seq_num
}

func (this *Package) GetMethod() string {
	return this.method
}
func (this *Package) CallError(error) bool {
	return false
}
func (this *Package) Marshal() ([]byte, error) {
	b := make([]byte, 0, 1 + 2 + 1 + 4 + 1 + 1 + len(this.method))
	b = append(b, 0xF8)
	b = append(b, byte(0), byte(0))
	b = append(b, this.rpc_type)
	b = append(b, byte(this.seq_num >> 24))
	b = append(b, byte(this.seq_num >> 16 & 0xff))
	b = append(b, byte(this.seq_num >> 8 & 0xff))
	b = append(b, byte(this.seq_num & 0xff))
	b = append(b, this.rsp_status)
	b = append(b, byte(len(this.method)))
	b = append(b, this.method...)
	if this.body != nil {
		body_b, e := proto.Marshal(this.body)
		if e != nil {
			return []byte{}, e
		}

		b = append(b, body_b...)
    }

	length := len(b)
	if length > 0xffff {
		return []byte{}, common.NewError(common.RpcError_PackTooLarge)
    }
	b[1] = byte(length >> 8)
	b[2] = byte(length & 0xff)
	return b, nil
}
func Unmarshal_PackageHead(b []byte) (pkg *Package, body []byte, consume int, err error) {
	if len(b) <= 5 {
		return 
    }

	if b[0] != 0xF8 {
		err = common.NewError(common.RpcError_MagicCodeError)
		return 
    }

	consume = int(b[1]) << 8 + int(b[2])
	if len(b) < consume {
		return 
	}

	pkg = &Package{}
	pkg.rpc_type = b[3]
	pkg.seq_num = uint32(b[4]) << 24 + uint32(b[5]) << 16 + uint32(b[6]) << 8 + uint32(b[7])
	pkg.rsp_status = b[8]
	head_tail := int(b[9]) + 10
	pkg.method = string(b[10:head_tail])
	body = b[head_tail:consume]
	return 
}

func Unmarshal_Body(pkg *Package, b []byte) (err error) {
	e := proto.Unmarshal(b, pkg.body)
	if e != nil {
		err = common.NewErrorMsg(int(common.RpcError_ParseError), e.Error())
	}

	return 
}

func (this *Package) GetBody() proto.Message {
	return this.body
}
/// ------------------------- 

/// ------------------------- Context
func NewContext(pkg *Package, conn IRpcConn) *Context {
	return &Context{*pkg, conn}
}

func (this *Context) Reply(status byte, response proto.Message) error {
	return this.conn.reply(this, status, response)
}
/// -------------------------

/// -------------------------  CallContext
func NewCallContext(method string, body proto.Message, cb RpcCallback) *CallContext {
	rpc_type := RpcType_Request
	if cb == nil { rpc_type = RpcType_Oneway }
	return &CallContext{*NewRequestPackage(rpc_type, method, body), false, cb}
}
func (this *CallContext) Call(err error, response proto.Message) bool {
	if this.called == true {
		return false
    }

	this.called = true
	go this.cb(err, response)
	return true
}
func (this *CallContext) CallError(err error) bool {
	return this.Call(err, nil)
}
/// ------------------------- 

