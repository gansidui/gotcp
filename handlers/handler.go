package handlers

import (
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/gansidui/gotcp/packet"
	"github.com/gansidui/gotcp/protomsgs"
	"github.com/gansidui/gotcp/utils"
	"log"
	"net"
	"time"
)

func SendByteStream(conn *net.TCPConn, buf []byte, writeTimeout time.Duration) error {
	conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	n, err := conn.Write(buf)
	if n != len(buf) || err != nil {
		return fmt.Errorf("Write to %v failed, Error: %v", conn.RemoteAddr(), err)
	}
	return nil
}

func HandleChatMsg(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	readmsg := &protomsgs.ChatMsg{}
	proto.Unmarshal(recPacket.Data, readmsg)

	fmt.Println(readmsg.GetSenderName())
	fmt.Println(readmsg.GetMsg())
	fmt.Println(utils.TimestampToTimestring(readmsg.GetTimestamp()))

	// write
	writemsg := &protomsgs.ChatMsg{
		SenderName: proto.String("服务器"),
		Msg:        proto.String("hello,生活真美好."),
		Timestamp:  proto.Int64(time.Now().Unix()),
	}
	data, err := proto.Marshal(writemsg)
	if err != nil {
		log.Printf("proto.Marshal(writemsg) error: %v", err)
		return
	}

	pac := &packet.Packet{
		Len:  uint32(len(data) + 6),
		Type: packet.TYPE_MESSAGE,
		Data: data,
	}

	SendByteStream(conn, pac.GetBytes(), 5*time.Second)
}
