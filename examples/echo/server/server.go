package main

import (
	"fmt"
	"github.com/gansidui/gotcp"
	"github.com/gansidui/gotcp/examples/echo"
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
	addr := c.GetRawConn().RemoteAddr()
	fmt.Println("OnConnect:", addr)
	c.PutExtraData(addr)
	return true
}

func (this *ConnDelegate) OnMessage(c *gotcp.Conn, p *gotcp.Packet) bool {
	fmt.Println("OnMessage:", p.GetLen(), p.GetTypeInt(), string(p.GetData()))
	c.AsyncWritePacket(echo.NewPacket(200, []byte("reply ok")), 5*time.Second)
	return true
}

func (this *ConnDelegate) OnClose(c *gotcp.Conn) {
	if extraData := c.GetExtraData(); extraData != nil {
		fmt.Println("OnClose:", c.GetExtraData())
	}

}

func (this *ConnDelegate) OnIOError(c *gotcp.Conn, err error) {
	if extraData := c.GetExtraData(); extraData != nil {
		fmt.Println("OnIOError:", c.GetExtraData(), err)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// listen
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	// set config and delegate
	config := &gotcp.Config{
		AcceptTimeout:          10 * time.Second,
		ReadTimeout:            120 * time.Second,
		WriteTimeout:           120 * time.Second,
		MaxPacketLength:        2048,
		SendPacketChanLimit:    10,
		ReceivePacketChanLimit: 10,
	}
	delegate := &ConnDelegate{}
	protocol := &echo.LtvProtocol{}

	// start server
	svr := gotcp.NewServer(config, delegate, protocol)
	go svr.Start(listener)

	// catch signal
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Signal: %v\r\n", <-ch)

	// stop server
	svr.Stop()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
