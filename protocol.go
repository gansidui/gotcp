package gotcp

import (
	"io"
)

type Packet interface {
	Serialize() []byte
}

type Protocol interface {
	ReadPacket(r io.Reader, packetSizeLimit uint32) (Packet, error)
}
