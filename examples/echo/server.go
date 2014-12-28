package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gansidui/gotcp"
	"github.com/gansidui/gotcp/protocol"
)

type Callback struct{}

func (this *Callback) OnConnect(c *gotcp.Conn) bool {
	addr := c.GetRawConn().RemoteAddr()
	c.PutExtraData(addr)
	fmt.Println("OnConnect:", addr)

	return true
}

func (this *Callback) OnMessage(c *gotcp.Conn, p gotcp.Packet) bool {
	ltvPacket := p.(*protocol.LtvPacket)

	fmt.Println("OnMessage:", ltvPacket.GetLen(), ltvPacket.GetType(),
		string(ltvPacket.GetValue()))

	if ltvPacket.GetType() == 88 {
		fmt.Println("bye bye", c.GetExtraData())
		return false
	}

	c.AsyncWritePacket(protocol.NewLtvPacket(200, []byte("welcome")), 0)

	return true
}

func (this *Callback) OnClose(c *gotcp.Conn) {
	fmt.Println("OnClose:", c.GetExtraData())
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// create a listener
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	// initialize server params
	config := &gotcp.Config{
		AcceptTimeout:          5 * time.Second,
		ReadTimeout:            240 * time.Second,
		WriteTimeout:           240 * time.Second,
		PacketSizeLimit:        2048,
		PacketSendChanLimit:    20,
		PacketReceiveChanLimit: 20,
	}
	srv := gotcp.NewServer(config, &Callback{}, &protocol.LtvProtocol{})

	// start server
	go srv.Start(listener)
	fmt.Println("listening:", listener.Addr())

	// catch system signal
	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Signal: ", <-chSig)

	// stop server
	srv.Stop()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
