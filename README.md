Gotcp is a powerful package for quickly writing tcp applications/services in golang

Install the gotcp package
~~~
go get github.com/gansidui/gotcp
~~~

Create server.go file
~~~ go
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

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	config := &gotcp.Config{
		AcceptTimeout:  30 * time.Second,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxPacLen:      uint32(1024),
		RecPacBufLimit: uint32(20),
	}

	callbacks := &gotcp.Callbacks{
		OnConnect:       onConnect,
		OnDisconnect:    onDisconnect,
		OnReceivePacket: onReceivePacket,
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

<<<<<<< HEAD
	svr := gotcp.NewServer(config, callbacks)
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

func onConnect(conn *net.TCPConn) error {
	fmt.Println("onConnect", conn.RemoteAddr())
	return nil
}

func onDisconnect(conn *net.TCPConn) {
	fmt.Println("onDisconnect", conn.RemoteAddr())
}

func onReceivePacket(conn *net.TCPConn, pac *gotcp.Packet) error {
	fmt.Println("onReceivePacket", conn.RemoteAddr())
	fmt.Println(pac.GetLen(), pac.GetType(), string(pac.GetData()))
	return nil
}
~~~

Run server
~~~
go run server.go
~~~
=======
释放socket占用的内存。


####下面这个chatserver是在gotcp的基础上写的，已经用于生产环境：https://github.com/gansidui/chatserver
>>>>>>> 9ca583c7c52c9f1f0c58403eef8dc93272d545d5
