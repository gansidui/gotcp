package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gansidui/gotcp/protocol"
)

func main() {
	// connect to server
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	ltvProtocol := &protocol.LtvProtocol{}

	// ping <--> pong
	for i := 0; i < 3; i++ {
		conn.Write(protocol.NewLtvPacket(123, []byte("hello")).Serialize())

		p, err := ltvProtocol.ReadPacket(conn, 1024)
		if err == nil {
			ltvPacket := p.(*protocol.LtvPacket)
			fmt.Println("Server reply:", ltvPacket.GetLen(), ltvPacket.GetType(),
				string(ltvPacket.GetValue()))
		}

		time.Sleep(2 * time.Second)
	}

	// bye bye
	conn.Write(protocol.NewLtvPacket(88, []byte("hello")).Serialize())

	time.Sleep(5 * time.Second)

	conn.Close()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
