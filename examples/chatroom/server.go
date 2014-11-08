package main

import (
	"fmt"
	"github.com/gansidui/go-utils/safemap"
	"github.com/gansidui/gotcp"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// all message packet type
const (
	TYPE_LOGIN = iota + 1
	TYPE_LOGOUT
	TYPE_MSG

	TYPE_REPLY_LOGIN
	TYPE_REPLY_LOGOUT
	TYPE_REPLY_MSG
)

type ConnDelegate struct {
	clientConns *safemap.SafeMap // save all client connection (conn --> bool)
}

func NewConnDelegate() *ConnDelegate {
	return &ConnDelegate{
		clientConns: safemap.New(),
	}
}

func (this *ConnDelegate) OnConnect(c *gotcp.Conn) bool {
	// read the first packet
	p, err := c.AsyncReadPacket(5 * time.Second)

	// check err and pacType, the first packet must be a login packet
	if err != nil || p.GetType() != TYPE_LOGIN {
		fmt.Println("Not received the client login packet:", err)
		return false
	}

	// save data
	uuid := string(p.GetData())
	c.PutExtraData(uuid)
	this.clientConns.Set(c, true)

	// reply
	c.WritePacket(gotcp.NewPacket(TYPE_REPLY_LOGIN, []byte("Login OK")))

	fmt.Printf("OnConnect:[%v]\n", uuid)
	return true
}

func (this *ConnDelegate) OnMessage(c *gotcp.Conn, p *gotcp.Packet) bool {
	fmt.Printf("OnMessage:[%v][%v,%v]\n", c.GetExtraData().(string), p.GetType(), string(p.GetData()))

	if p.GetType() == TYPE_LOGOUT {
		c.AsyncWritePacket(gotcp.NewPacket(TYPE_REPLY_LOGOUT, []byte("Logout OK")), 5*time.Second)
		return false

	} else if p.GetType() == TYPE_MSG {

		// send message to all client connections
		clients := this.clientConns.Items()
		for conn, _ := range clients {
			if conn != c {
				conn.(*gotcp.Conn).AsyncWritePacket(gotcp.NewPacket(TYPE_MSG, p.GetData()), 5*time.Second)
			}
		}

		// reply
		c.AsyncWritePacket(gotcp.NewPacket(TYPE_REPLY_MSG, []byte("Msg OK")), 5*time.Second)
	}

	return true
}

func (this *ConnDelegate) OnClose(c *gotcp.Conn) {
	if extraData := c.GetExtraData(); extraData != nil {
		fmt.Println("OnClose:", c.GetExtraData())
	}
	this.clientConns.Delete(c)
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
	delegate := NewConnDelegate()

	// start server
	svr := gotcp.NewServer(config, delegate)
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
