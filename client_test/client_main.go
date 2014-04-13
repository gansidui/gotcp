package main

import (
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/gansidui/gotcp/handlers"
	"github.com/gansidui/gotcp/packet"
	"github.com/gansidui/gotcp/protomsgs"
	"github.com/gansidui/gotcp/utils"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

func test(id int) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	fmt.Println("Connected:", id)

	ticker := time.NewTicker(3 * time.Second)
	for _ = range ticker.C {
		// write
		writemsg := &protomsgs.ChatMsg{
			SenderName: proto.String("客户端: " + strconv.Itoa(id)),
			Msg:        proto.String("你好，丫的，hello world!!!!"),
			Timestamp:  proto.Int64(time.Now().Unix()),
		}

		data, _ := proto.Marshal(writemsg)
		pac := &packet.Packet{
			Len:  uint32(len(data) + 6),
			Type: packet.TYPE_MESSAGE,
			Data: data,
		}

		// err := handlers.SendByteStream(conn, pac.GetBytes())
		// if err != nil {
		// 	log.Printf("Error: %v", err)
		// }

		// 测试： 分段发送
		dd := pac.GetBytes()
		handlers.SendByteStream(conn, dd[0:2], 5*time.Second)
		time.Sleep(1 * time.Second)
		handlers.SendByteStream(conn, dd[2:7], 5*time.Second)
		time.Sleep(1 * time.Second)
		handlers.SendByteStream(conn, dd[7:], 5*time.Second)

		// read
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		checkError(err)

		readmsg := &protomsgs.ChatMsg{}
		proto.Unmarshal(buf[6:n], readmsg)

		fmt.Println(readmsg.GetSenderName())
		fmt.Println(readmsg.GetMsg())
		fmt.Println(utils.TimestampToTimestring(readmsg.GetTimestamp()))
	}

}

func main() {
	for i := 0; i < 40000; i++ {
		time.Sleep(50 * time.Millisecond)
		go test(i)
	}
	time.Sleep(3600 * time.Second)
}

func checkError(err error) {
	if err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
