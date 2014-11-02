package gotcp

import (
	"errors"
)

// Error types
var (
	ConnIsClosedError   = errors.New("Conn is closed")
	WriteIsBlockedError = errors.New("Write packet is blocked")
	ReadIsBlockedError  = errors.New("Read packet is blocked")
	ReadPacketError     = errors.New("Read packet error")
	PacketTooLargeError = errors.New("Packet too large")
)

// BigEndian: int32 --> []byte
func Int32ToBytes(v int32) []byte {
	b := make([]byte, 4)
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
	return b
}

// BigEndian: []byte -->int32
func BytesToInt32(b []byte) int32 {
	return int32(b[3]) | int32(b[2])<<8 | int32(b[1])<<16 | int32(b[0])<<24
}
