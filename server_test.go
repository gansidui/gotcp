package gotcp

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// test tips
/******************************************************/

var OnConnectTip, OnMessageTip, OnCloseTip, OnIOErrorTip string

/******************************************************/

// delegate
/******************************************************/
type Delegate struct{}

func (this *Delegate) OnConnect(c *Conn) bool {
	p, err := c.AsyncReadPacket(5 * time.Second)
	if err != nil {
		return false
	}

	OnConnectTip = fmt.Sprintf("OnConnect[%v,%v,%v]", p.GetLen(), p.GetType(), string(p.GetData()))

	fmt.Println(OnConnectTip)
	return true
}

func (this *Delegate) OnMessage(c *Conn, p *Packet) bool {
	OnMessageTip = fmt.Sprintf("OnMessage[%v,%v,%v]", p.GetLen(), p.GetType(), string(p.GetData()))
	fmt.Println(OnMessageTip)

	if string(p.GetData()) == "logout" {
		c.WritePacket(NewPacket(888, []byte("ok")))
		return false
	}

	c.AsyncWritePacket(NewPacket(999, []byte(string(p.GetData())+",ok")), 5*time.Second)

	return true
}

func (this *Delegate) OnClose(c *Conn) {
	OnCloseTip = fmt.Sprintf("OnClose[%v]", c.IsClosed())
	fmt.Println(OnCloseTip)
}

func (this *Delegate) OnIOError(c *Conn, err error) {
	if err != nil {
		OnIOErrorTip = fmt.Sprintf("OnIOError[%v]", err)
	}
	fmt.Println(OnIOErrorTip)
}

/******************************************************/

func simulateClient(t *testing.T) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)

	// OnConnect
	conn.Write(NewPacket(777, []byte("login")).Serialize())
	time.Sleep(100 * time.Millisecond)
	if OnConnectTip != "OnConnect[13,777,login]" {
		t.Fatal()
	}

	// OnMessage
	conn.Write(NewPacket(666, []byte("helloworld")).Serialize())
	time.Sleep(100 * time.Millisecond)
	if OnMessageTip != "OnMessage[18,666,helloworld]" {
		t.Fatal()
	}

	retPacket, _ := ReadPacket(conn, 2048)
	if retPacket.GetLen() != 21 || retPacket.GetType() != 999 || string(retPacket.GetData()) != "helloworld,ok" {
		t.Fatal()
	}

	// OnClose
	conn.Write(NewPacket(555, []byte("logout")).Serialize())
	time.Sleep(100 * time.Millisecond)
	if OnMessageTip != "OnMessage[14,555,logout]" {
		t.Fatal()
	}

	retPacket, _ = ReadPacket(conn, 2048)
	if retPacket.GetLen() != 10 || retPacket.GetType() != 888 || string(retPacket.GetData()) != "ok" {
		t.Fatal()
	}

	if OnCloseTip != "OnClose[true]" {
		t.Fatal()
	}

	// OnIOError
	if OnIOErrorTip != fmt.Sprintf("OnIOError[%v]", ReadPacketError) {
		t.Fatal()
	}
}

func TestServer(t *testing.T) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	if err != nil {
		t.Fatal()
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		t.Fatal()
	}

	config := &Config{
		AcceptTimeout:          5 * time.Second,
		ReadTimeout:            5 * time.Second,
		WriteTimeout:           5 * time.Second,
		MaxPacketLength:        2048,
		SendPacketChanLimit:    10,
		ReceivePacketChanLimit: 10,
	}
	delegate := &Delegate{}

	svr := NewServer(config, delegate)
	go svr.Start(listener)

	time.Sleep(time.Second)

	simulateClient(t)

	svr.Stop()
}
