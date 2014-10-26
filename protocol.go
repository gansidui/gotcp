package gotcp

import (
	"io"
)

// Packet: pacLen + pacType + pacData
// Big endian: int32 + int32 + []byte
type Packet struct {
	pacLen  int32
	pacType int32
	pacData []byte
}

func NewPacket(pacType int32, pacData []byte) *Packet {
	return &Packet{
		pacLen:  int32(8) + int32(len(pacData)),
		pacType: pacType,
		pacData: pacData,
	}
}

func (p *Packet) Serialize() []byte {
	buf := make([]byte, 8+len(p.pacData))
	copy(buf[0:4], Int32ToBytes(p.pacLen))
	copy(buf[4:8], Int32ToBytes(p.pacType))
	copy(buf[8:], p.pacData)
	return buf
}

func (p *Packet) GetLen() int32 {
	return p.pacLen
}

func (p *Packet) GetType() int32 {
	return p.pacType
}

func (p *Packet) GetData() []byte {
	return p.pacData
}

func ReadPacket(r io.Reader, MaxPacketLength int32) (*Packet, error) {
	var (
		pacBLen  []byte = make([]byte, 4)
		pacBType []byte = make([]byte, 4)
		pacLen   int32
	)

	// read pacLen
	if n, err := io.ReadFull(r, pacBLen); err != nil && n != 4 {
		return nil, ReadPacketError
	}
	if pacLen = BytesToInt32(pacBLen); pacLen > MaxPacketLength {
		return nil, PacketTooLargeError
	}

	// read pacType
	if n, err := io.ReadFull(r, pacBType); err != nil && n != 4 {
		return nil, ReadPacketError
	}

	// read pacData
	pacData := make([]byte, pacLen-8)
	if n, err := io.ReadFull(r, pacData); err != nil && n != int(pacLen) {
		return nil, ReadPacketError
	}

	return NewPacket(BytesToInt32(pacBType), pacData), nil
}
