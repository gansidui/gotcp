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

	LfpProtocol := &protocol.LfpProtocol{}

	// ping <--> pong
	for i := 0; i < 3; i++ {
		conn.Write(protocol.NewLfpPacket([]byte("hello")).Serialize())

		p, err := LfpProtocol.ReadPacket(conn, 1024)
		if err == nil {
			lfpPacket := p.(*protocol.LfpPacket)
			message := lfpPacket.GetBuffer()
			fmt.Println("Server reply:", string(message))
		}

		time.Sleep(2 * time.Second)
	}

	// bye bye
	conn.Write(protocol.NewLfpPacket([]byte("bye")).Serialize())

	time.Sleep(5 * time.Second)

	conn.Close()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
