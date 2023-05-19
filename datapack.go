package wss

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const MESSAGE_HEAD_LEN = 8

type IDataPack interface {
	GetHeadLen() uint32
	Pack(msg IMessage) ([]byte, error)
	Unpack([]byte) (IMessage, error)
}

type DataPack struct{}

func NewDataPack() *DataPack {
	return &DataPack{}
}

func (dp *DataPack) GetHeadLen() uint32 {
	return MESSAGE_HEAD_LEN
}

func (dp *DataPack) Pack(msg IMessage) ([]byte, error) {
	dataBuff := bytes.NewBuffer([]byte{})

	if err := binary.Write(dataBuff, binary.BigEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}

	if err := binary.Write(dataBuff, binary.BigEndian, msg.GetMsgID()); err != nil {
		return nil, err
	}

	if err := binary.Write(dataBuff, binary.BigEndian, msg.GetData()); err != nil {
		return nil, err
	}
	return dataBuff.Bytes(), nil
}

func (dp *DataPack) Unpack(bs []byte) (IMessage, error) {
	dataBuff := bytes.NewReader(bs)
	msg := &Message{}

	if err := binary.Read(dataBuff, binary.BigEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	if err := binary.Read(dataBuff, binary.BigEndian, &msg.ID); err != nil {
		return nil, err
	}

	if DefaultMaxPacketSize > 0 && msg.DataLen > DefaultMaxPacketSize {
		return nil, errors.New("too large msg data received")
	}

	return msg, nil
}
