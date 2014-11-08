package main

import (
	"errors"
	"fmt"
	utils "github.com/gansidui/go-utils"
	"github.com/gansidui/gotcp"
	"log"
	"net"
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

func main() {
	conn, err := connect()
	checkError(err)
	defer conn.Close()

	sendLogin(conn)

	// read data
	go func(conn *net.TCPConn) {
		for {
			p, err := gotcp.ReadPacket(conn, 2048)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(p.GetLen(), p.GetType(), string(p.GetData()))
		}

	}(conn)

	// send data
	exitChan := make(chan struct{}, 0)
	time.AfterFunc(10*time.Second, func() { exitChan <- struct{}{} })

	for {
		select {
		case <-exitChan:
			sendLogout(conn)
			fmt.Println("=======================")
			goto exit

		default:
			sendMsg(conn)
			time.Sleep(time.Second)
		}
	}

exit:
	fmt.Println("BYE BYE")
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func connect() (*net.TCPConn, error) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	return net.DialTCP("tcp", nil, tcpAddr)
}

func sendLogin(conn *net.TCPConn) error {
	uuid := "client===" + utils.RandomString(3, 5) + "==="
	fmt.Println("uuid===", uuid)

	conn.Write(gotcp.NewPacket(TYPE_LOGIN, []byte(uuid)).Serialize())
	p, err := gotcp.ReadPacket(conn, 2048)
	if err != nil {
		return err
	}

	if p.GetType() != TYPE_REPLY_LOGIN || string(p.GetData()) != "Login OK" {
		return errors.New("LOGIN FAILED")
	}

	fmt.Println(p.GetLen(), p.GetType(), string(p.GetData()))

	return nil
}

func sendLogout(conn *net.TCPConn) {
	conn.Write(gotcp.NewPacket(TYPE_LOGOUT, []byte("BYE BYE")).Serialize())
}

func sendMsg(conn *net.TCPConn) {
	conn.Write(gotcp.NewPacket(TYPE_MSG, []byte("hello world "+utils.RandomString(3, 5))).Serialize())
}
