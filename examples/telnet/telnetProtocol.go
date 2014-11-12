package telnet

import (
	"fmt"
	"io"
	//"time"

	"github.com/gansidui/gotcp"
)

// Packet
type TelnetPacketDelegate struct {
	pLen  uint32
	pType uint32
	pData []byte
}

func (p *TelnetPacketDelegate) Serialize() []byte {
	return p.pData
}

func (p *TelnetPacketDelegate) GetLen() uint32 {
	return p.pLen
}

func (p *TelnetPacketDelegate) GetTypeInt() uint32 {
	return p.pType
}

func (p *TelnetPacketDelegate) GetTypeString() string {
	return ""
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

type MosProtocol struct {
}

func converUtf16ToUtf8(sourceData []byte) []byte {
	utils.InformationalWithFormat("原始mos_Data:% X", sourceData)
	d := mahonia.NewDecoder("UTF-16")
	if d == nil {
		utils.Critical("Could not create decoder for UTF-16")
		return nil
	}

	_, data, err := d.Translate(sourceData, true)
	if err != nil {
		utils.Critical(err.Error())
		return nil
	}
	return data
}

func (this *MosProtocol) ReadPacket(r io.Reader, MaxPacketLength uint32) (*gotcp.Packet, error) {
	data := make([]byte, MaxPacketLength) // 设定缓存空间

	readLengh, err := r.Read(data)

	if err != nil { //EOF, or worse
		return nil, err
	}

	if readLengh == 0 { // 连接可能已被客户端关闭
		return nil, gotcp.ReadPacketError
	} else {
		//return NewPacket(0, data[0:readLengh]), nil
		return NewPacket(0, converUtf16ToUtf8(data[0:readLengh])), nil
	}
}

type MosConnDelegate struct {
	connectCount int
	closeCount   int
	messageCount int
}

func (this *MosConnDelegate) OnConnect(c *gotcp.Conn) bool {
	this.connectCount++
	c.PutExtraData(this.connectCount)

	fmt.Printf("OnConnect[%s][***%v***]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int))
	return true
}

func (this *MosConnDelegate) OnMessage(c *gotcp.Conn, p *gotcp.Packet) bool {
	fmt.Printf("OnMessage[%s][***%v***]:[%v]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int), string(p.GetData()))
	this.messageCount++
	command := string(p.GetData())
	utils.Informational("mos:", command)
	if !utf8.Valid(p.GetData()) {
		utils.Informational("指令编码无效")
		return true
	}
	//utils.Informational("valid utf8:", utf8.Valid(p.GetData()))
	utils.InformationalWithFormat("mos_Data:% X", p.GetData())

	command = "<mymos>" + command + "</mymos>"

	DoWork(command)
	return true
}

func (this *MosConnDelegate) OnClose(c *gotcp.Conn) {
	this.closeCount++
	fmt.Printf("OnClose[%s][***%v***]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int))
}

func (this *MosConnDelegate) OnIOError(c *gotcp.Conn, err error) {
	fmt.Printf("OnIOError[%s][***%v***]:[%v]\n", c.GetRawConn().RemoteAddr(), c.GetExtraData().(int), err)
}
