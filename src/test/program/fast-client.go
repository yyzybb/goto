package main

import (
	"net"
	"time"
	"runtime"
)
import "goto_rpc"
import airth "airth"
import "fmt"
import proto "encoding/protobuf/proto"

func main() {
	go func() {
		c := time.Tick(time.Millisecond * 1000)
		for {
			select { case <-c: }
			fmt.Println("goroutine count:", runtime.NumGoroutine())
		}
	}()

	goto_rpc.CloseLog()

	// startup client
	conn, e := net.Dial("tcp", "127.0.0.1:8090")
	if e != nil {
		fmt.Println("connect to 8090 error!", e.Error())
		return
	}

	fmt.Println("client connected...")
	client := goto_rpc.NewClient(conn, 3000000)
	stub, e := airth.NewArithService_Stub(client)
	if e != nil {
		fmt.Println("init stub error!", e.Error())
		return
	}

	req := &airth.ArithRequest{proto.Int(5), proto.Int(6), nil}

	count := 30000 * 10
	recv := 0
	exit_c := make(chan int)
	start := time.Now()
	for i := 0; i < count; i++ {
		e = stub.AsynMultiply(req, func(e error, response *airth.ArithResponse) {
			if e != nil {
				fmt.Println("AsyncMultiply go call error:", e.Error())
			} else {
				//fmt.Println("asyn rpc response: ", response.GetVal())
			}

			recv ++
			if recv == count {
				exit_c <- 1
            }
		})
		if e != nil {
			fmt.Println("rpc call error!", e.Error())
		}

		runtime.Gosched()
	}

	<-exit_c;
	fmt.Println("done")
	end := time.Now()
	fmt.Println(end.Sub(start))
}
