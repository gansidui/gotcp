package gotcp

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// test tips
/******************************************************/

var OnConnectInfo, OnMessageInfo, OnCloseInfo, OnIOErrorInfo string

/******************************************************/

// delegate
/******************************************************/
type Delegate struct{}

func (this *Delegate) OnConnect(c *Conn) bool {
	p, err := c.AsyncReadPacket(5 * time.Second)
	if err != nil {
		return false
	}

	OnConnectInfo = fmt.Sprintf("OnConnect[%v,%v,%v]", p.GetLen(), p.GetType(), string(p.GetData()))

	fmt.Printf("OnConnect[%v,%v,%v]\n", p.GetLen(), p.GetType(), string(p.GetData()))
	return true
}

func (this *Delegate) OnMessage(c *Conn, p *Packet) bool {
	OnMessageInfo = fmt.Sprintf("OnMessage[%v,%v,%v]", p.GetLen(), p.GetType(), string(p.GetData()))
	fmt.Println(OnMessageInfo)

	if string(p.GetData()) == "logout" {
		c.WritePacket(NewPacket(888, []byte("ok")))
		return false
	}

	c.AsyncWritePacket(NewPacket(999, []byte(string(p.GetData())+",ok")), 5*time.Second)

	return true
}

func (this *Delegate) OnClose(c *Conn) {
	OnCloseInfo = fmt.Sprintf("OnClose[%v]", c.IsClosed())
	fmt.Println(OnCloseInfo)
}

func (this *Delegate) OnIOError(c *Conn, err error) {
	if err != nil {
		OnIOErrorInfo = fmt.Sprintf("OnIOError[%v]", err)
	}
	fmt.Println(OnIOErrorInfo)
}

/******************************************************/

func simulateClient(t *testing.T) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)

	// OnConnect
	conn.Write(NewPacket(777, []byte("login")).Serialize())
	time.Sleep(100 * time.Millisecond)
	if OnConnectInfo != "OnConnect[13,777,login]" {
		t.Fatal()
	}

	// OnMessage
	conn.Write(NewPacket(666, []byte("helloworld")).Serialize())
	time.Sleep(100 * time.Millisecond)
	if OnMessageInfo != "OnMessage[18,666,helloworld]" {
		t.Fatal()
	}

	retPacket, _ := ReadPacket(conn, 2048)
	if retPacket.GetLen() != 21 || retPacket.GetType() != 999 || string(retPacket.GetData()) != "helloworld,ok" {
		t.Fatal()
	}

	// OnClose
	conn.Write(NewPacket(555, []byte("logout")).Serialize())
	time.Sleep(100 * time.Millisecond)
	if OnMessageInfo != "OnMessage[14,555,logout]" {
		t.Fatal()
	}

	retPacket, _ = ReadPacket(conn, 2048)
	if retPacket.GetLen() != 10 || retPacket.GetType() != 888 || string(retPacket.GetData()) != "ok" {
		t.Fatal()
	}

	if OnCloseInfo != "OnClose[true]" {
		t.Fatal()
	}

	// OnIOError
	if OnIOErrorInfo != fmt.Sprintf("OnIOError[%v]", ReadPacketError) {
		t.Fatal()
	}
}

func TestServer(t *testing.T) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	listener, _ := net.ListenTCP("tcp", tcpAddr)

	config := &Config{
		AcceptTimeout:          5 * time.Second,
		ReadTimeout:            5 * time.Second,
		WriteTimeout:           5 * time.Second,
		MaxPacketLength:        int32(2048),
		SendPacketChanLimit:    int32(10),
		ReceivePacketChanLimit: int32(10),
	}
	delegate := &Delegate{}

	svr := NewServer(config, delegate)
	go svr.Start(listener)

	time.Sleep(time.Second)

	simulateClient(t)

	svr.Stop()
}
