package gotcp_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/gansidui/gotcp"
	"github.com/gansidui/gotcp/examples/echo"
)

// Server delegate
type ServerDelegate struct {
	t *testing.T
}

func (this *ServerDelegate) OnConnect(c *gotcp.Conn) bool {
	fmt.Println("Server OnConnect")
	return true
}

func (this *ServerDelegate) OnMessage(c *gotcp.Conn, p *gotcp.Packet) bool {
	fmt.Println("Server OnMessage")

	if p.GetTypeInt() == 777 {
		if string(p.GetData()) != "BYE" {
			this.t.Fatal()
		}
		return false
	}

	if p.GetTypeInt() != 999 || string(p.GetData()) != "hello" {
		this.t.Fatal()
	}

	c.WritePacket(echo.NewPacket(888, []byte("world")))

	return true
}

func (this *ServerDelegate) OnClose(c *gotcp.Conn) {
	fmt.Println("Server OnClose")
}

func (this *ServerDelegate) OnIOError(c *gotcp.Conn, err error) {
	fmt.Println("Server OnIOError")
}

// Client delegate
type ClientDelegate struct {
	t *testing.T
}

func (this *ClientDelegate) OnConnect(c *gotcp.Conn) bool {
	fmt.Println("Client OnConnect")
	return true
}

func (this *ClientDelegate) OnMessage(c *gotcp.Conn, p *gotcp.Packet) bool {
	fmt.Println("Client OnMessage")

	if p.GetTypeInt() != 888 || string(p.GetData()) != "world" {
		this.t.Fatal()
	}

	c.AsyncWritePacket(echo.NewPacket(777, []byte("BYE")), time.Second)

	return true
}

func (this *ClientDelegate) OnClose(c *gotcp.Conn) {
	fmt.Println("Client OnClose")
}

func (this *ClientDelegate) OnIOError(c *gotcp.Conn, err error) {
	fmt.Println("Client OnIOError")
}

func TestDial(t *testing.T) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8990")
	if err != nil {
		t.Fatal()
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		t.Fatal()
	}

	config := &gotcp.Config{
		AcceptTimeout:          5 * time.Second,
		ReadTimeout:            5 * time.Second,
		WriteTimeout:           5 * time.Second,
		MaxPacketLength:        2048,
		SendPacketChanLimit:    10,
		ReceivePacketChanLimit: 10,
	}
	delegate := &ServerDelegate{t: t}
	protocol := &echo.LtvProtocol{}

	svr := gotcp.NewServer(config, delegate, protocol)
	go svr.Start(listener)

	time.Sleep(time.Second)

	simulateClientDial(t, svr)

	svr.Stop()
}

func simulateClientDial(t *testing.T, svr *gotcp.Server) {
	config := &gotcp.Config{
		AcceptTimeout:          5 * time.Second,
		ReadTimeout:            5 * time.Second,
		WriteTimeout:           5 * time.Second,
		MaxPacketLength:        2048,
		SendPacketChanLimit:    10,
		ReceivePacketChanLimit: 10,
	}
	delegate := &ClientDelegate{t: t}
	protocol := &echo.LtvProtocol{}

	conn, err := svr.Dial("tcp4", "127.0.0.1:8990", config, delegate, protocol)
	if err != nil {
		t.Fatal()
	}

	go conn.Do()
	time.Sleep(time.Second)

	conn.WritePacket(echo.NewPacket(999, []byte("hello")))
	time.Sleep(time.Second)
}
