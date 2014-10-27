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

const (
	TYPE_LOGIN = iota + 1
	TYPE_LOGOUT
	TYPE_MSG

	TYPE_REPLY_LOGIN
	TYPE_REPLY_LOGOUT
	TYPE_REPLY_MSG
)

type ConnDelegate struct {
	connectCount int
	closeCount   int
	messageCount int
}

func (this *ConnDelegate) OnConnect(c *gotcp.Conn) bool {
	p, err := c.AsyncReadPacket(5 * time.Second)
	if err != nil {
		fmt.Printf("OnConnect[Error]:[%v]\n", err)
		return false
	}

	if p.GetType() == TYPE_LOGIN && string(p.GetData()) == "LOGIN" {
		c.WritePacket(gotcp.NewPacket(TYPE_REPLY_LOGIN, []byte("LOGIN OK")))

		this.connectCount++
		c.PutExtraData(this.connectCount)

		fmt.Printf("OnConnect[LOGIN][***%v***]\n", c.GetExtraData().(int))
		return true
	}

	return false
}

func (this *ConnDelegate) OnMessage(c *gotcp.Conn, p *gotcp.Packet) bool {
	if p.GetType() == TYPE_LOGOUT {
		c.WritePacket(gotcp.NewPacket(TYPE_REPLY_LOGOUT, []byte("LOGOUT OK")))
		fmt.Printf("OnMessage[LOGOUT][***%v***]:[%v]\n", c.GetExtraData().(int), string(p.GetData()))
		return false
	}

	if p.GetType() == TYPE_MSG {
		c.AsyncWritePacket(gotcp.NewPacket(TYPE_REPLY_MSG, []byte("REPLY_"+string(p.GetData()))), 5*time.Second)

		this.messageCount++

		fmt.Printf("OnMessage[MSG][***%v***]:[%v]\n", c.GetExtraData().(int), string(p.GetData()))
		return true
	}

	return true
}

func (this *ConnDelegate) OnClose(c *gotcp.Conn) {
	this.closeCount++
	fmt.Printf("OnClose[***%v***]\n", c.GetExtraData().(int))
}

func (this *ConnDelegate) OnIOError(c *gotcp.Conn, err error) {
	fmt.Printf("OnIOError[***%v***]:[%v]\n", c.GetExtraData().(int), err)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	config := &gotcp.Config{
		AcceptTimeout:          10 * time.Second,
		ReadTimeout:            120 * time.Second,
		WriteTimeout:           120 * time.Second,
		MaxPacketLength:        int32(2048),
		SendPacketChanLimit:    int32(10),
		ReceivePacketChanLimit: int32(10),
	}
	delegate := &ConnDelegate{}

	svr := gotcp.NewServer(config, delegate)
	go svr.Start(listener)

	go func() {
		for {
			fmt.Println("===========goroutine==========", runtime.NumGoroutine())
			fmt.Println(delegate.connectCount, delegate.closeCount, delegate.messageCount)
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
