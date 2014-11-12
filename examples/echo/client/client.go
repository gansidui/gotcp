package main

import (
	"fmt"
	"github.com/gansidui/gotcp/examples/echo"
	"log"
	"net"
	"time"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	ltvProtocol := &echo.LtvProtocol{}

	for i := 0; i < 3; i++ {
		conn.Write(echo.NewPacket(999, []byte("hello world")).Serialize())

		p, err := ltvProtocol.ReadPacket(conn, 1024)
		if err == nil {
			fmt.Println(p.GetLen(), p.GetTypeInt(), string(p.GetData()))
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
