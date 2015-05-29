package airth

import proto "encoding/protobuf/proto"
import json "encoding/json"
import math "math"
import "goto_rpc"

import "io"
import "net"
import "net/rpc"
import "net/rpc/protorpc"

// Reference proto, json, and math imports to suppress error if they are not otherwise used.
var _ = proto.Marshal
var _ = &json.SyntaxError{}
var _ = math.Inf

type ArithRequest struct {
	A                *int32 `protobuf:"varint,1,opt,name=a" json:"a,omitempty"`
	B                *int32 `protobuf:"varint,2,opt,name=b" json:"b,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *ArithRequest) Reset()         { *m = ArithRequest{} }
func (m *ArithRequest) String() string { return proto.CompactTextString(m) }
func (*ArithRequest) ProtoMessage()    {}

func (m *ArithRequest) GetA() int32 {
	if m != nil && m.A != nil {
		return *m.A
	}
	return 0
}

func (m *ArithRequest) GetB() int32 {
	if m != nil && m.B != nil {
		return *m.B
	}
	return 0
}

type ArithResponse struct {
	Val              *int32 `protobuf:"varint,1,opt,name=val" json:"val,omitempty"`
	Quo              *int32 `protobuf:"varint,2,opt,name=quo" json:"quo,omitempty"`
	Rem              *int32 `protobuf:"varint,3,opt,name=rem" json:"rem,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *ArithResponse) Reset()         { *m = ArithResponse{} }
func (m *ArithResponse) String() string { return proto.CompactTextString(m) }
func (*ArithResponse) ProtoMessage()    {}

func (m *ArithResponse) GetVal() int32 {
	if m != nil && m.Val != nil {
		return *m.Val
	}
	return 0
}

func (m *ArithResponse) GetQuo() int32 {
	if m != nil && m.Quo != nil {
		return *m.Quo
	}
	return 0
}

func (m *ArithResponse) GetRem() int32 {
	if m != nil && m.Rem != nil {
		return *m.Rem
	}
	return 0
}

func init() {
}

type IArithServiceAsyn interface {
	Multiply(ctx goto_rpc.IContext, request *ArithRequest)
	Divide(ctx goto_rpc.IContext, request *ArithRequest)
}

func RegisterArithServiceAsyn(srv *goto_rpc.Server, service IArithServiceAsyn) error {
	e := srv.AddServiceFunc("ArithService.Multiply", func(ctx goto_rpc.IContext, request proto.Message) {
		service.Multiply(ctx, request.(*ArithRequest))
    }, func() proto.Message {
		return &ArithRequest{}
    }, func() proto.Message {
		return &ArithResponse{}
	})
	if e != nil { return e }

	e = srv.AddServiceFunc("ArithService.Divide", func(ctx goto_rpc.IContext, request proto.Message) {
		service.Divide(ctx, request.(*ArithRequest))
    }, func() proto.Message {
		return &ArithRequest{}
    }, func() proto.Message {
		return &ArithResponse{}
	})
	if e != nil { return e }

	return nil
}

type IArithServiceSync interface {
	Multiply(ctx goto_rpc.IContext, request *ArithRequest) (response *ArithResponse, status byte)
	Divide(ctx goto_rpc.IContext, request *ArithRequest) (response *ArithResponse, status byte)
}

func RegisterArithServiceSync(srv *goto_rpc.Server, service IArithServiceSync) error {
	e := srv.AddServiceFunc("ArithService.Multiply", func(ctx goto_rpc.IContext, request proto.Message) {
		rsp, s := service.Multiply(ctx, request.(*ArithRequest))
		ctx.Reply(s, rsp)
    }, func() proto.Message {
		return &ArithRequest{}
    }, func() proto.Message {
		return &ArithResponse{}
	})
	if e != nil { return e }

	e = srv.AddServiceFunc("ArithService.Divide", func(ctx goto_rpc.IContext, request proto.Message) {
		rsp, s := service.Divide(ctx, request.(*ArithRequest))
		ctx.Reply(s, rsp)
    }, func() proto.Message {
		return &ArithRequest{}
    }, func() proto.Message {
		return &ArithResponse{}
	})
	if e != nil { return e }

	return nil
}

type ArithService_Stub struct {
	*goto_rpc.Client
}

func NewArithService_Stub(c *goto_rpc.Client) (stub *ArithService_Stub, e error) {
	e = c.AddServiceInfo("ArithService.Multiply",
	func() proto.Message {
		return &ArithRequest{}
	}, func() proto.Message {
		return &ArithResponse{}
	})
	if e != nil { return }

	e = c.AddServiceInfo("ArithService.Divide",
	func() proto.Message {
		return &ArithRequest{}
	}, func() proto.Message {
		return &ArithResponse{}
	})
	if e != nil { return }
	stub = &ArithService_Stub{c}
	return
}

func (stub *ArithService_Stub) Multiply(request *ArithRequest) (*ArithResponse, error) {
	rsp, e := stub.Call("ArithService.Multiply", request)
	response, _ := rsp.(*ArithResponse)
	return response, e
}
func (stub *ArithService_Stub) Divide(request *ArithRequest) (*ArithResponse, error) {
	rsp, e := stub.Call("ArithService.Divide", request)
	response, _ := rsp.(*ArithResponse)
	return response, e
}

func (stub *ArithService_Stub) AsynMultiply(request *ArithRequest, cb func(error, *ArithResponse)) {
	stub.AsynCall("ArithService.Multiply", request, func(err error, rsp proto.Message) {
		response, _ := rsp.(*ArithResponse)
		cb(err, response)
    })
}
func (stub *ArithService_Stub) AsynDivide(request *ArithRequest, cb func(error, *ArithResponse)) {
	stub.AsynCall("ArithService.Divide", request, func(err error, rsp proto.Message) {
		response, _ := rsp.(*ArithResponse)
		cb(err, response)
    })
}







type ArithService interface {
	Multiply(in *ArithRequest, out *ArithResponse) error
	Divide(in *ArithRequest, out *ArithResponse) error
}

// RegisterArithService publish the given ArithService implementation on the server.
func RegisterArithService(srv *rpc.Server, x ArithService) error {
	if err := srv.RegisterName("ArithService", x); err != nil {
		return err
	}
	return nil
}

// ServeArithService serves the given ArithService implementation on conn.
func ServeArithService(conn io.ReadWriteCloser, x ArithService) error {
	srv := rpc.NewServer()
	if err := srv.RegisterName("ArithService", x); err != nil {
		return err
	}
	srv.ServeCodec(protorpc.NewServerCodec(conn))
	return nil
}

// ListenAndServeArithService listen announces on the local network address laddr
// and serves the given ArithService implementation.
func ListenAndServeArithService(network, addr string, x ArithService) error {
	clients, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	srv := rpc.NewServer()
	if err := srv.RegisterName("ArithService", x); err != nil {
		return err
	}
	for {
		conn, err := clients.Accept()
		if err != nil {
			return err
		}
		go srv.ServeCodec(protorpc.NewServerCodec(conn))
	}
	panic("unreachable")
}

type rpcArithServiceStub struct {
	*rpc.Client
}

func (c *rpcArithServiceStub) Multiply(in *ArithRequest, out *ArithResponse) error {
	return c.Call("ArithService.Multiply", in, out)
}
func (c *rpcArithServiceStub) Divide(in *ArithRequest, out *ArithResponse) error {
	return c.Call("ArithService.Divide", in, out)
}

// DialArithService connects to an ArithService at the specified network address.
func DialArithService(network, addr string) (*rpc.Client, ArithService, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, nil, err
	}
	c, srv := NewArithServiceClient(conn)
	return c, srv, nil
}

// NewArithServiceClient returns a ArithService rpc.Client and stub to handle
// requests to the set of ArithService at the other end of the connection.
func NewArithServiceClient(conn io.ReadWriteCloser) (*rpc.Client, ArithService) {
	c := rpc.NewClientWithCodec(protorpc.NewClientCodec(conn))
	return c, &rpcArithServiceStub{c}
}

// NewArithServiceStub returns a ArithService stub to handle rpc.Client.
func NewArithServiceStub(c *rpc.Client) ArithService {
	return &rpcArithServiceStub{c}
}
