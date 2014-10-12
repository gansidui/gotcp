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
			tcpAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				log.Fatal(err)
			}

			pac := gotcp.NewPacket(211314, []byte("hello world"))
			conn.Write(pac.Serialize())
			conn.Close()

			fmt.Println("j====", j)

		}(j)

		time.Sleep(20 * time.Millisecond)
	}

	time.Sleep(30 * time.Second)
}
