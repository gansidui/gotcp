package main

import (
	"fmt"
	"github.com/gansidui/gotcp"
	"log"
	"net"
	"time"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	for i := 0; i < 3; i++ {
		conn.Write(gotcp.NewPacket(999, []byte("hello world")).Serialize())

		p, err := gotcp.ReadPacket(conn, 1024)
		if err == nil {
			fmt.Println(p.GetLen(), p.GetType(), string(p.GetData()))
		}

		time.Sleep(3 * time.Second)
	}

	conn.Close()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
