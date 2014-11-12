package gotcp

import (
	"io"
)

type PacketDelegate interface {
	Serialize() []byte
	GetLen() uint32
	GetData() []byte
	GetTypeInt() uint32
	GetTypeString() string
}

type Packet struct {
	length   int32
	Delegate PacketDelegate
}

func (p *Packet) Serialize() []byte {
	return p.Delegate.Serialize()
}

func (p *Packet) GetLen() uint32 {
	return p.Delegate.GetLen()
}

func (p *Packet) GetData() []byte {
	return p.Delegate.GetData()
}

func (p *Packet) GetTypeInt() uint32 {
	return p.Delegate.GetTypeInt()
}

func (p *Packet) GetTypeString() string {
	return p.Delegate.GetTypeString()
}

type Protocol interface {
	ReadPacket(r io.Reader, MaxPacketLength uint32) (*Packet, error)
}
