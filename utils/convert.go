package utils

import (
	"time"
)

// 均采用大端字节序，事实上这个字节序没有任何用处，
// 只要服务器和客户端约定采用相同的字节序就行

func Uint32ToBytes(v uint32) []byte {
	buf := make([]byte, 4)
	buf[0] = byte(v >> 24)
	buf[1] = byte(v >> 16)
	buf[2] = byte(v >> 8)
	buf[3] = byte(v)
	return buf
}

func Uint16ToBytes(v uint16) []byte {
	buf := make([]byte, 2)
	buf[0] = byte(v >> 8)
	buf[1] = byte(v)
	return buf
}

func BytesToUint32(buf []byte) uint32 {
	v := (uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3]))
	return v
}

func BytesToUint16(buf []byte) uint16 {
	v := (uint16(buf[0])<<8 | uint16(buf[1]))
	return v
}

func TimestampToTimestring(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}
