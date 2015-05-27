package test

import "testing"
import "net"
import "goto_rpc"
import airth "airth"
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
	// initialize server.
	lstn, e := net.Listen("tcp", "127.0.0.1:8090")
	if e != nil {
		t.Fatal("listen 8090 error!", e.Error())
		return 
	}
	fmt.Println("src initialize ok.")

	srv := goto_rpc.NewServer(lstn)
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
	client := goto_rpc.NewClient(conn)
	stub := airth.NewArithService_Stub(client)

	var a int32 = 5
	var b int32 = 6
	req := &airth.ArithRequest{&a, &b, nil}
	fmt.Println("rpc call...")
	rsp, e := stub.Multiply(req)
	if e != nil {
		t.Fatal("rpc call error!", e.Error())
		return 
	}

	fmt.Println("rpc response: ", rsp.GetVal())
}
