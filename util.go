package gotcp

import (
	"errors"
)

// Errors
var (
	ConnIsClosedError   = errors.New("Conn is closed")
	WriteIsBlockedError = errors.New("Write packet is blocked")
	ReadPacketError     = errors.New("Read packet error")
	PacketTooLargeError = errors.New("Packet too large")
)

// Convert int32 to []byte
func Int32ToBytes(v int32) []byte {
	buf := make([]byte, 4)
	buf[0] = byte(v >> 24)
	buf[1] = byte(v >> 16)
	buf[2] = byte(v >> 8)
	buf[3] = byte(v)
	return buf
}

// Convert []byte to int32
func BytesToInt32(buf []byte) int32 {
	v := (int32(buf[0])<<24 | int32(buf[1])<<16 | int32(buf[2])<<8 | int32(buf[3]))
	return v
}
