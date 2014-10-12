package gotcp

/*

packet: pacLen + pacType + pacData
big endian: uint32 + uint32 + []byte

*/

type Packet struct {
	pacLen  uint32
	pacType uint32
	pacData []byte
}

func NewPacket(pacType uint32, pacData []byte) *Packet {
	return &Packet{
		pacLen:  uint32(8) + uint32(len(pacData)),
		pacType: pacType,
		pacData: pacData,
	}
}

func (p *Packet) Serialize() []byte {
	buf := make([]byte, 0)
	buf = append(buf, Uint32ToBytes(p.pacLen)...)
	buf = append(buf, Uint32ToBytes(p.pacType)...)
	buf = append(buf, p.pacData...)
	return buf
}

func (p *Packet) GetLen() uint32 {
	return p.pacLen
}

func (p *Packet) GetType() uint32 {
	return p.pacType
}

func (p *Packet) GetData() []byte {
	return p.pacData
}
