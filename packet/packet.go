package packet

import (
	"github.com/gansidui/gotcp/utils"
)

// 每个数据包的组成： 包长 + 类型 + 数据
//               -->  4字节 + 2字节 + 数据
//               -->  uint32 + uint16 + []byte
// 数据部分用protobuf封装

// 数据包的类型
const (
	TYPE_LOGIN = uint16(iota)
	TYPE_LOGOUT
	TYPE_MESSAGE
	TYPE_PINT
)

type Packet struct {
	Len  uint32
	Type uint16
	Data []byte
}

// 序列化Packet
func (this *Packet) GetBytes() (buf []byte) {
	buf = append(buf, utils.Uint32ToBytes(this.Len)...)
	buf = append(buf, utils.Uint16ToBytes(this.Type)...)
	buf = append(buf, this.Data...)
	return buf
}
