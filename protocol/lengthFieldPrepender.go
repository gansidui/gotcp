package protocol

import (
	"io"

	"github.com/gansidui/gotcp"
)

// packet: the length value is prepended as a binary form. (length field prepender)
// total length = len(buffer) = lengthBytes + body
type LfpPacket struct {
	buffer []byte
}

func (this *LfpPacket) Serialize() []byte {
	return this.buffer
}

func (this *LfpPacket) GetLength() uint32 {
	return BytesToUint32(this.buffer[0:4])
}

func (this *LfpPacket) GetBody() []byte {
	return this.buffer[4:]
}

func NewLfpPacket(buffer []byte, lenFieldFlag bool) *LfpPacket {
	pac := &LfpPacket{}

	if lenFieldFlag {
		pac.buffer = buffer

	} else {
		pac.buffer = make([]byte, 4+len(buffer))
		copy(pac.buffer[0:4], Uint32ToBytes(uint32(len(buffer))))
		copy(pac.buffer[4:], buffer)
	}

	return pac
}

type LfpProtocol struct {
}

func (this *LfpProtocol) ReadPacket(r io.Reader, packetSizeLimit uint32) (gotcp.Packet, error) {
	var (
		lengthBytes []byte = make([]byte, 4)
		length      uint32
	)

	// read length
	if _, err := io.ReadFull(r, lengthBytes); err != nil {
		return nil, ErrReadPacket
	}
	if length = BytesToUint32(lengthBytes); length > packetSizeLimit {
		return nil, ErrPacketTooLarger
	}

	// read body ( buffer = lengthBytes + body )
	buffer := make([]byte, 4+length)
	copy(buffer[0:4], lengthBytes)
	if _, err := io.ReadFull(r, buffer[4:]); err != nil {
		return nil, ErrReadPacket
	}

	return NewLfpPacket(buffer, true), nil
}
