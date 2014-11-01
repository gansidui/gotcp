package main

import (
	"errors"
	"fmt"
	"github.com/gansidui/gotcp"
	"log"
	"net"
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

func main() {
	for j := 0; j < 100000; j++ {

		go func(j int) {
			conn, err := connect()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			fmt.Println("connect  ====== ", j)

			if err = sendLogin(conn); err != nil {
				log.Fatal(err)
			}

			for i := 0; i < 200; i++ {
				time.Sleep(6 * time.Second)
				if err = sendMsg(conn); err != nil {
					log.Fatal(err)
				}
			}

			if err = sendLogout(conn); err != nil {
				log.Fatal(err)
			}

			fmt.Println("disconnect  ****** ", j)

		}(j)

		time.Sleep(20 * time.Millisecond)
	}

	time.Sleep(time.Hour)
}

func connect() (*net.TCPConn, error) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	return net.DialTCP("tcp", nil, tcpAddr)
}

func sendLogin(conn *net.TCPConn) error {
	conn.Write(gotcp.NewPacket(TYPE_LOGIN, []byte("LOGIN")).Serialize())
	p, err := gotcp.ReadPacket(conn, 2048)
	if err != nil {
		return err
	}

	if p.GetType() != TYPE_REPLY_LOGIN || string(p.GetData()) != "LOGIN OK" {
		return errors.New("LOGIN FAILED")
	}

	return nil
}

func sendLogout(conn *net.TCPConn) error {
	conn.Write(gotcp.NewPacket(TYPE_LOGOUT, []byte("BYE BYE")).Serialize())
	p, err := gotcp.ReadPacket(conn, 2048)
	if err != nil {
		return err
	}

	if p.GetType() != TYPE_REPLY_LOGOUT || string(p.GetData()) != "LOGOUT OK" {
		return errors.New("LOGOUT FAILED")
	}

	return nil
}

func sendMsg(conn *net.TCPConn) error {
	conn.Write(gotcp.NewPacket(TYPE_MSG, []byte("hello world")).Serialize())
	p, err := gotcp.ReadPacket(conn, 2048)
	if err != nil {
		return err
	}

	if p.GetType() != TYPE_REPLY_MSG || string(p.GetData()) != "REPLY_hello world" {
		return errors.New("MSG FAILED")
	}

	return nil
}
