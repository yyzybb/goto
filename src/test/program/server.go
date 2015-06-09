package main

import "net"
import "goto_rpc"
import airth "test/airth"
import "fmt"
import "time"
import (
	"os"
	"os/signal"
	"runtime/pprof"
)

type ArithServiceAsyn struct { }

var request_c uint64 = 0

func (this *ArithServiceAsyn) Multiply(ctx goto_rpc.IContext, request *airth.ArithRequest) {
	request_c++
	//fmt.Println("On Multiply")
	val := request.GetA() * request.GetB()
	response := &airth.ArithResponse{}
	response.Val = &val
	e := ctx.GoReply(0, response)
	if e != nil {
		fmt.Println("reply error: ", e.Error())
    }
}

func (this *ArithServiceAsyn) Divide(ctx goto_rpc.IContext, request *airth.ArithRequest) {
	request_c++
	//fmt.Println("On Divide")
	val := request.GetA() / request.GetB()
	response := &airth.ArithResponse{}
	response.Val = &val
	e := ctx.GoReply(0, response)
	if e != nil {
		fmt.Println("reply error: ", e.Error())
    }
}

func main() {
	//pprof
	f, _ := os.Create("profile_server")
	pprof.StartCPUProfile(f)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		<-c
		pprof.StopCPUProfile()
		os.Exit(0)
    }()

	fmt.Println("start test")
	go func() {
		c := time.Tick(time.Second * 1)
		for {
			select {
			case <-c:
            }

			fmt.Println("request_c:", request_c)
		}
    }()
	server()
}

func server() {
	goto_rpc.CloseLog()

	// initialize server.
	lstn, e := net.Listen("tcp", "127.0.0.1:8090")
	if e != nil {
		fmt.Println("listen 8090 error!", e.Error())
		return 
	}
	fmt.Println("src initialize ok.")

	srv := goto_rpc.NewServer(lstn, 3000000)
	airth_service := &ArithServiceAsyn{}
	e = airth.RegisterArithServiceAsyn(srv, airth_service)
	if e != nil {
		fmt.Println("Register error!", e.Error())
		return 
	}

	fmt.Println("srv started...")
	srv.Start()
}
