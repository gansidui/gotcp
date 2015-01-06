package protocol

import (
	"io"

	"github.com/gansidui/gotcp"
)

// Packet: pacLen + pacType + pacValue
// BigEndian: uint32 + uint32 + []byte
type LtvPacket struct {
	pacLen   uint32
	pacType  uint32
	pacValue []byte
}

func (p *LtvPacket) Serialize() []byte {
	buf := make([]byte, 8+len(p.pacValue))
	copy(buf[0:4], Uint32ToBytes(p.pacLen))
	copy(buf[4:8], Uint32ToBytes(p.pacType))
	copy(buf[8:], p.pacValue)
	return buf
}

func (p *LtvPacket) GetLen() uint32 {
	return p.pacLen
}

func (p *LtvPacket) GetType() uint32 {
	return p.pacType
}

func (p *LtvPacket) GetValue() []byte {
	return p.pacValue
}

func NewLtvPacket(pacType uint32, pacValue []byte) *LtvPacket {
	return &LtvPacket{
		pacLen:   uint32(8) + uint32(len(pacValue)),
		pacType:  pacType,
		pacValue: pacValue,
	}
}

type LtvProtocol struct {
}

func (this *LtvProtocol) ReadPacket(r io.Reader, packetSizeLimit uint32) (gotcp.Packet, error) {
	var (
		pacLenBytes  []byte = make([]byte, 4)
		pacTypeBytes []byte = make([]byte, 4)
		pacLen       uint32
	)

	// read pacLen
	if _, err := io.ReadFull(r, pacLenBytes); err != nil {
		return nil, ErrReadPacket
	}
	if pacLen = BytesToUint32(pacLenBytes); pacLen > packetSizeLimit {
		return nil, ErrPacketTooLarger
	}

	// read pacType
	if _, err := io.ReadFull(r, pacTypeBytes); err != nil {
		return nil, ErrReadPacket
	}

	// read pacValue
	pacValue := make([]byte, pacLen-8)
	if _, err := io.ReadFull(r, pacValue); err != nil {
		return nil, ErrReadPacket
	}

	return NewLtvPacket(BytesToUint32(pacTypeBytes), pacValue), nil
}
