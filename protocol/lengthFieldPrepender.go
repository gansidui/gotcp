package protocol

import (
	"io"

	"github.com/gansidui/gotcp"
)

// packet: the length value is prepended as a binary form. (length field prepender)
// total length = 2 + len(buffer) = 2 + length
type LfpPacket struct {
	length uint16
	buffer []byte
}

func (this *LfpPacket) Serialize() []byte {
	buf := make([]byte, 2+len(this.buffer))
	copy(buf[0:2], Uint16ToBytes(this.length))
	copy(buf[2:], this.buffer)
	return buf
}

func (this *LfpPacket) GetBuffer() []byte {
	return this.buffer
}

func NewLfpPacket(buffer []byte) *LfpPacket {
	return &LfpPacket{
		length: uint16(len(buffer)),
		buffer: buffer,
	}
}

type LfpProtocol struct {
}

func (this *LfpProtocol) ReadPacket(r io.Reader, packetSizeLimit uint32) (gotcp.Packet, error) {
	var (
		lengthBytes []byte = make([]byte, 2)
		length      uint16
	)

	// read length
	if _, err := io.ReadFull(r, lengthBytes); err != nil {
		return nil, ErrReadPacket
	}
	if length = BytesToUint16(lengthBytes); uint32(length) > packetSizeLimit {
		return nil, ErrPacketTooLarger
	}

	// read buffer
	buffer := make([]byte, length)
	if _, err := io.ReadFull(r, buffer); err != nil {
		return nil, ErrReadPacket
	}

	return NewLfpPacket(buffer), nil
}
