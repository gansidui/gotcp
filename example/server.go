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

// count
var connectNum, closeNum, messageNum, ioErrorNum int

type ConnDelegate struct{}

func (this *ConnDelegate) OnConnect(c *gotcp.Conn) bool {
	connectNum++

	c.SetReadDeadline(5 * time.Second)
	if p, err := c.ReadPacket(); err == nil {
		fmt.Println("OnConnect:", p.GetLen(), p.GetType(), string(p.GetData()))
	}

	return true
}

func (this *ConnDelegate) OnMessage(c *gotcp.Conn, p *gotcp.Packet) bool {
	fmt.Println("OnMessage:", p.GetLen(), p.GetType(), string(p.GetData()))
	messageNum++

	c.SetWriteDeadline(5 * time.Second)
	c.WritePacket(gotcp.NewPacket(99, []byte("hello")))

	c.AsyncWritePacket(gotcp.NewPacket(100, []byte("world")), 5*time.Second)

	return true
}

func (this *ConnDelegate) OnClose(c *gotcp.Conn) {
	closeNum++
	fmt.Println("OnClose")
}

func (this *ConnDelegate) OnIOError(c *gotcp.Conn, err error) {
	ioErrorNum++
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

	go func() {
		for {
			fmt.Println("=======num goroutine === ", runtime.NumGoroutine())
			fmt.Println(connectNum, closeNum, messageNum)
			time.Sleep(2 * time.Second)
		}
	}()

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
