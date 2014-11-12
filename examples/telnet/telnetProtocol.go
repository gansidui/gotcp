package telnet

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gansidui/gotcp"
)

// Packet
type TelnetPacketDelegate struct {
	pLen        uint32
	pType       uint32
	pTypeString string
	pData       []byte
}

func (p *TelnetPacketDelegate) Serialize() []byte {
	buf := p.pData
	endTag := []byte("\r\n")
	buf = append(buf, endTag...)
	return buf
}

func (p *TelnetPacketDelegate) GetLen() uint32 {
	return p.pLen
}

func (p *TelnetPacketDelegate) GetTypeInt() uint32 {
	return p.pType
}

func (p *TelnetPacketDelegate) GetTypeString() string {
	return p.pTypeString
}

func (p *TelnetPacketDelegate) GetData() []byte {
	return p.pData
}

func NewPacket(pType uint32, pData []byte) *gotcp.Packet {
	packet := new(gotcp.Packet)
	packet.Delegate = &TelnetPacketDelegate{
		pLen:  uint32(len(pData)),
		pType: pType,
		pData: pData,
	}
	return packet
}

func NewPacket2(pTypeString string, pData []byte) *gotcp.Packet {
	packet := new(gotcp.Packet)
	packet.Delegate = &TelnetPacketDelegate{
		pLen:        uint32(len(pData)),
		pTypeString: pTypeString,
		pData:       pData,
	}
	return packet
}

type TelnetProtocol struct {
}

func (this *TelnetProtocol) ReadPacket(r io.Reader, MaxPacketLength uint32) (*gotcp.Packet, error) {
	fullBuf := bytes.NewBuffer([]byte{})
	endTag := []byte("\r\n") //Telnet command's end tag
	for {
		data := make([]byte, MaxPacketLength)

		readLengh, err := r.Read(data)

		if err != nil { //EOF, or worse
			return nil, err
		}

		if readLengh == 0 { // Connection maybe closed by the client
			return nil, gotcp.ConnClosedError
		} else {
			fullBuf.Write(data[:readLengh])

			index := bytes.Index(fullBuf.Bytes(), endTag)
			if index > -1 {
				command := fullBuf.Next(index)
				fullBuf.Next(2)
				//fmt.Println(string(command))

				commandList := strings.Split(string(command), " ")
				if len(commandList) > 1 {
					return NewPacket2(commandList[0], []byte(commandList[1])), nil
				} else {
					return NewPacket2("unknow", command), nil
				}
			}
		}
	}
}

type TelnetConnDelegate struct {
	connectCount int
	closeCount   int
	messageCount int
}

func (this *TelnetConnDelegate) OnConnect(c *gotcp.Conn) bool {
	this.connectCount++
	c.PutExtraData(this.connectCount)
	fmt.Printf("OnConnect[%s][***%v***]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int))

	c.WritePacket(NewPacket(1, []byte("Welcome to this Telnet Server")))
	return true
}

func (this *TelnetConnDelegate) OnMessage(c *gotcp.Conn, p *gotcp.Packet) bool {
	fmt.Printf("OnMessage[%s][***%v***]:[%v]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int), string(p.GetData()))
	this.messageCount++
	command := p.GetData()
	typeStr := p.GetTypeString()
	switch typeStr {
	case "echo":
		c.WritePacket(NewPacket(1, command))
	case "login":
		c.WritePacket(NewPacket(2, []byte(string(command)+" has login")))
	default:
		c.WritePacket(NewPacket(0, []byte("unknow command")))
	}

	return true
}

func (this *TelnetConnDelegate) OnClose(c *gotcp.Conn) {
	this.closeCount++
	fmt.Printf("OnClose[%s][***%v***]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int))
}

func (this *TelnetConnDelegate) OnIOError(c *gotcp.Conn, err error) {
	fmt.Printf("OnIOError[%s][***%v***]:[%v]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int), err)
}
