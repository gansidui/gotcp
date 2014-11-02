package gotcp

import (
	"encoding/binary"
	"errors"
)

// Error types
var (
	ConnClosedError     = errors.New("connection was closed")
	WriteBlockedError   = errors.New("write blocking")
	ReadBlockedError    = errors.New("read blocking")
	ReadPacketError     = errors.New("read packet error")
	PacketTooLargeError = errors.New("packet too large")
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
