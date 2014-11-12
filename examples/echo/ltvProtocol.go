package echo

import (
	"io"

	"github.com/gansidui/gotcp"
)

// Packet: pacLen + pacType + pacData
// Big endian: int32 + int32 + []byte
type LtvPacketDelegate struct {
	pacLen  uint32
	pacType uint32
	pacData []byte
}

func (p *LtvPacketDelegate) Serialize() []byte {
	buf := make([]byte, 8+len(p.pacData))
	copy(buf[0:4], gotcp.Uint32ToBytes(p.pacLen))
	copy(buf[4:8], gotcp.Uint32ToBytes(p.pacType))
	copy(buf[8:], p.pacData)
	return buf
}

func (p *LtvPacketDelegate) GetLen() uint32 {
	return p.pacLen
}

func (p *LtvPacketDelegate) GetTypeInt() uint32 {
	return p.pacType
}

func (p *LtvPacketDelegate) GetTypeString() string {
	return ""
}

func (p *LtvPacketDelegate) GetData() []byte {
	return p.pacData
}

func NewPacket(pacType uint32, pacData []byte) *gotcp.Packet {
	packet := new(gotcp.Packet)
	packet.Delegate = &LtvPacketDelegate{
		pacLen:  uint32(8) + uint32(len(pacData)),
		pacType: pacType,
		pacData: pacData,
	}
	return packet
}

type LtvProtocol struct {
}

func (this *LtvProtocol) ReadPacket(r io.Reader, MaxPacketLength uint32) (*gotcp.Packet, error) {
	var (
		pacBLen  []byte = make([]byte, 4)
		pacBType []byte = make([]byte, 4)
		pacLen   uint32
	)

	// read pacLen
	if n, err := io.ReadFull(r, pacBLen); err != nil && n != 4 {
		return nil, gotcp.ReadPacketError
	}
	if pacLen = gotcp.BytesToUint32(pacBLen); pacLen > MaxPacketLength {
		return nil, gotcp.PacketTooLargeError
	}

	// read pacType
	if n, err := io.ReadFull(r, pacBType); err != nil && n != 4 {
		return nil, gotcp.ReadPacketError
	}

	// read pacData
	pacData := make([]byte, pacLen-8)
	if n, err := io.ReadFull(r, pacData); err != nil && n != int(pacLen) {
		return nil, gotcp.ReadPacketError
	}

	return NewPacket(gotcp.BytesToUint32(pacBType), pacData), nil
}
