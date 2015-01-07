package telnet

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gansidui/gotcp"
)

var (
	endTag = []byte("\r\n") //Telnet command's end tag
)

// Packet
type TelnetPacket struct {
	pLen  uint32
	pType string
	pData []byte
}

func (p *TelnetPacket) Serialize() []byte {
	buf := p.pData
	buf = append(buf, endTag...)
	return buf
}

func (p *TelnetPacket) GetLen() uint32 {
	return p.pLen
}

func (p *TelnetPacket) GetType() string {
	return p.pType
}

func (p *TelnetPacket) GetData() []byte {
	return p.pData
}

func NewTelnetPacket(pType string, pData []byte) *TelnetPacket {
	return &TelnetPacket{
		pLen:  uint32(len(pData)),
		pType: pType,
		pData: pData,
	}
}

type TelnetProtocol struct {
}

func (this *TelnetProtocol) ReadPacket(r io.Reader, packetSizeLimit uint32) (gotcp.Packet, error) {
	fullBuf := bytes.NewBuffer([]byte{})
	for {
		data := make([]byte, packetSizeLimit)

		readLengh, err := r.Read(data)

		if err != nil { //EOF, or worse
			return nil, err
		}

		if readLengh == 0 { // Connection maybe closed by the client
			return nil, gotcp.ErrConnClosing
		} else {
			fullBuf.Write(data[:readLengh])

			index := bytes.Index(fullBuf.Bytes(), endTag)
			if index > -1 {
				command := fullBuf.Next(index)
				fullBuf.Next(2)
				//fmt.Println(string(command))

				commandList := strings.Split(string(command), " ")
				if len(commandList) > 1 {
					return NewTelnetPacket(commandList[0], []byte(commandList[1])), nil
				} else {
					if commandList[0] == "quit" {
						return NewTelnetPacket("quit", command), nil
					} else {
						return NewTelnetPacket("unknow", command), nil
					}
				}
			}
		}
	}
}

type TelnetCallback struct {
	connectCount int
	closeCount   int
	messageCount int
}

func (this *TelnetCallback) OnConnect(c *gotcp.Conn) bool {
	this.connectCount++
	c.PutExtraData(this.connectCount)
	fmt.Printf("OnConnect[%s][***%v***]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int))

	c.AsyncWritePacket(NewTelnetPacket("unknow", []byte("Welcome to this Telnet Server")), 0)
	return true
}

func (this *TelnetCallback) OnMessage(c *gotcp.Conn, p gotcp.Packet) bool {
	packet := p.(*TelnetPacket)

	fmt.Printf("OnMessage[%s][***%v***]:[%v]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int), string(packet.GetData()))
	this.messageCount++
	command := packet.GetData()
	commandType := packet.GetType()
	switch commandType {
	case "echo":
		c.AsyncWritePacket(NewTelnetPacket("echo", command), 0)
	case "login":
		c.AsyncWritePacket(NewTelnetPacket("login", []byte(string(command)+" has login")), 0)
	case "quit":
		return false
	default:
		c.AsyncWritePacket(NewTelnetPacket("unknow", []byte("unknow command")), 0)
	}

	return true
}

func (this *TelnetCallback) OnClose(c *gotcp.Conn) {
	this.closeCount++
	fmt.Printf("OnClose[%s][***%v***]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int))
}

//func (this *TelnetConnDelegate) OnIOError(c *gotcp.Conn, err error) {
//	fmt.Printf("OnIOError[%s][***%v***]:[%v]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int), err)
//}
