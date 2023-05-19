package wss

import "fmt"

type IMessage interface {
	GetMsgID() uint32
	GetDataLen() uint32
	GetData() []byte

	SetMsgID(uint32)
	SetDataLen(uint32)
	SetData([]byte)
}

type Message struct {
	ID      uint32
	DataLen uint32
	Data    []byte
}

func NewMsgPackage(id uint32, data []byte) *Message {
	return &Message{
		ID:      id,
		DataLen: uint32(len(data)),
		Data:    data,
	}
}

func (msg *Message) GetMsgID() uint32 {
	return msg.ID
}

func (msg *Message) GetDataLen() uint32 {
	return msg.DataLen
}

func (msg *Message) GetData() []byte {
	return msg.Data
}

func (msg *Message) SetMsgID(id uint32) {
	msg.ID = id
}

func (msg *Message) SetDataLen(length uint32) {
	msg.DataLen = length
}

func (msg *Message) SetData(data []byte) {
	msg.Data = data
}

func (msg *Message) String() string {
	return fmt.Sprintf("{id: %d, data: %s}", msg.ID, string(msg.Data))
}
