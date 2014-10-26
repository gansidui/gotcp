package main

import (
	"fmt"
	"github.com/gansidui/gotcp"
	"log"
	"net"
	"time"
)

func main() {
	for j := 0; j < 100; j++ {

		go func(j int) {
			conn, err := connect()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			fmt.Println("connect  ====== ", j)

			conn.Write(gotcp.NewPacket(88, []byte("hi")).Serialize())

			ticker := time.NewTicker(3 * time.Second)
			for _ = range ticker.C {
				conn.Write(gotcp.NewPacket(211314, []byte("hello world")).Serialize())

				if pac, err := gotcp.ReadPacket(conn, 2048); err == nil {
					fmt.Println(pac.GetLen(), pac.GetType(), string(pac.GetData()))
				}

				if pac, err := gotcp.ReadPacket(conn, 2048); err == nil {
					fmt.Println(pac.GetLen(), pac.GetType(), string(pac.GetData()))
				}
			}

			fmt.Println("disconnect  ****** ", j)

		}(j)

		time.Sleep(20 * time.Millisecond)
	}

	time.Sleep(5 * time.Minute)
}

func connect() (*net.TCPConn, error) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	return net.DialTCP("tcp", nil, tcpAddr)
}
