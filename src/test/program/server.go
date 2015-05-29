package main

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

func main() {
	fmt.Println("start test")
	server()
}

func server() {
	// initialize server.
	lstn, e := net.Listen("tcp", "127.0.0.1:8090")
	if e != nil {
		fmt.Println("listen 8090 error!", e.Error())
		return 
	}
	fmt.Println("src initialize ok.")

	srv := goto_rpc.NewServer(lstn)
	airth_service := &ArithServiceAsyn{}
	e = airth.RegisterArithServiceAsyn(srv, airth_service)
	if e != nil {
		fmt.Println("Register error!", e.Error())
		return 
	}

	fmt.Println("srv started...")
	srv.Start()
}
