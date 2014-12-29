package protocol

import (
	"encoding/binary"
	"errors"
)

// error type
var (
	ErrReadPacket      = errors.New("read packet failed")
	ErrPacketTooLarger = errors.New("the size of packet is larger than the limit")
)

// uint32 --> []byte
func Uint32ToBytes(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

// []byte --> uint32
func BytesToUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

// uint16 --> []byte
func Uint16ToBytes(v uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	return b
}

// []byte --> uint16
func BytesToUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}
