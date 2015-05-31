package main

import (
	"net"
	"time"
)
import "goto_rpc"
import airth "airth"
import "fmt"

func main() {
	fmt.Println("start test")
	client()
	time.Sleep(5 * time.Second)
	fmt.Println("end")
}

func client() {
	// startup client
	conn, e := net.Dial("tcp", "127.0.0.1:8090")
	if e != nil {
		fmt.Println("connect to 8090 error!", e.Error())
		return
	}

	fmt.Println("client connected...")
	client := goto_rpc.NewClient(conn, 10)
	stub, e := airth.NewArithService_Stub(client)
	if e != nil {
		fmt.Println("init stub error!", e.Error())
		return
	}

	var a int32 = 5
	var b int32 = 6
	req := &airth.ArithRequest{&a, &b, nil}

	/*
		fmt.Println("rpc call...")
		rsp, e := stub.Multiply(req)
		if e != nil {
			fmt.Println("rpc call error!", e.Error())
			return
		}
		fmt.Println("sync rpc response: ", rsp.GetVal())
	*/

	for i := 0; i < 300; i++ {
		e = stub.GoAsynMultiply(req, func(e error, response *airth.ArithResponse) {
			if e != nil {
				fmt.Println("AsyncMultiply go call error:", e.Error())
			} else {
				fmt.Println("asyn rpc response: ", response.GetVal())
			}
		})
		if e != nil {
			fmt.Println("rpc call error!", e.Error())
		}
	}
}
