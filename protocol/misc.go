package protocol

import (
	"encoding/binary"
	"errors"
)

// simple error type
var (
	ErrReadPacket      = errors.New("read packet failed")
	ErrPacketTooLarger = errors.New("the size of packet is larger than the limit")
)

// BigEndian: uint32 --> []byte
func Uint32ToBytes(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

// BigEndian: []byte --> uint32
func BytesToUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}
