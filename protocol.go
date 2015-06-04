package gotcp

import (
	"net"
)

type Packet interface {
	Serialize() []byte
}

type Protocol interface {
	ReadPacket(conn *net.TCPConn) (Packet, error)
}
