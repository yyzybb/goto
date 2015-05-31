package test

import "testing"
import "net"
import "goto_rpc"
import airth "test/airth"
import proto "encoding/protobuf/proto"
import "fmt"

type ArithServiceAsyn struct { }

func (this *ArithServiceAsyn) Multiply(ctx goto_rpc.IContext, request *airth.ArithRequest) {
	fmt.Println("On Multiply")
	val := request.GetA() * request.GetB()
	response := &airth.ArithResponse{}
	response.Val = &val
	ctx.Reply(0, response)
}

func (this *ArithServiceAsyn) Divide(ctx goto_rpc.IContext, request *airth.ArithRequest) {
	fmt.Println("On Divide")
	val := request.GetA() / request.GetB()
	response := &airth.ArithResponse{}
	response.Val = &val
	ctx.Reply(0, response)
}

func TestServerAndClient(t *testing.T) {
	fmt.Println("start test")
	goto_rpc.CloseLog()
	// initialize server.
	lstn, e := net.Listen("tcp", "127.0.0.1:8090")
	if e != nil {
		t.Fatal("listen 8090 error!", e.Error())
		return 
	}
	fmt.Println("src initialize ok.")

	srv := goto_rpc.NewServer(lstn, 1)
	airth_service := &ArithServiceAsyn{}
	e = airth.RegisterArithServiceAsyn(srv, airth_service)
	if e != nil {
		t.Fatal("Register error!", e.Error())
		return 
	}

	go srv.Start()
	fmt.Println("srv started...")

	// startup client
	conn, e := net.Dial("tcp", "127.0.0.1:8090")
	if e != nil {
		t.Fatal("connect to 8090 error!", e.Error())
		return 
	}

	fmt.Println("client connected...")
	client := goto_rpc.NewClient(conn, 1)
	stub, _ := airth.NewArithService_Stub(client)

	req := &airth.ArithRequest{proto.Int(8), proto.Int(2), nil}
	fmt.Println("rpc call...")
	rsp, e := stub.Multiply(req)
	if e != nil {
		t.Fatal("rpc call error!", e.Error())
		return 
	}
	fmt.Println("rpc response: ", rsp.GetVal())

	rsp, e = stub.Divide(req)
	if e != nil {
		t.Fatal("rpc call error!", e.Error())
		return 
	}
	fmt.Println("rpc response: ", rsp.GetVal())
}
