gotcp
================

gotcp is a powerful package for quickly writing tcp applications/services in golang.


How to install
================

~~~
go get github.com/gansidui/gotcp
~~~

How to use
================

Create server.go file:

~~~go
package main

import (
	"fmt"
	"github.com/gansidui/gotcp"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type ConnDelegate struct{}

func (this *ConnDelegate) OnConnect(c *gotcp.Conn) bool {
	fmt.Println("OnConnect")
	return true
}

func (this *ConnDelegate) OnMessage(c *gotcp.Conn, p *gotcp.Packet) bool {
	fmt.Println("OnMessage:", p.GetLen(), p.GetType(), string(p.GetData()))
	return true
}

func (this *ConnDelegate) OnClose(c *gotcp.Conn) {
	fmt.Println("OnClose")
}

func (this *ConnDelegate) OnIOError(c *gotcp.Conn, err error) {
	fmt.Println("OnIOError:", err)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	config := &gotcp.Config{
		AcceptTimeout:          5 * time.Second,
		ReadTimeout:            5 * time.Second,
		WriteTimeout:           5 * time.Second,
		MaxPacketLength:        int32(2048),
		SendPacketChanLimit:    int32(10),
		ReceivePacketChanLimit: int32(10),
	}
	delegate := &ConnDelegate{}

	svr := gotcp.NewServer(config, delegate)
	go svr.Start(listener)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Signal: %v\r\n", <-ch)

	svr.Stop()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}


~~~


Run server:

~~~
go run server.go
~~~


Document
================

[Go Go](http://godoc.org/github.com/gansidui/gotcp)