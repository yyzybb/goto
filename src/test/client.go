package main

import "net"
import "goto_rpc"
import airth "airth"
import "fmt"

func main() {
	fmt.Println("start test")
	client()
}

func client() {
	// startup client
	conn, e := net.Dial("tcp", "127.0.0.1:8090")
	if e != nil {
		fmt.Println("connect to 8090 error!", e.Error())
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
		fmt.Println("rpc call error!", e.Error())
		return 
	}

	fmt.Println("rpc response: ", rsp.GetVal())
}
