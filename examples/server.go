package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/cjinle/wss"
)

func main() {
	ser := wss.NewServer("0.0.0.0:8080", "/", "wss")
	ser.SetMaxConn(10)
	ser.SetOnConnStart(OnConnectionAdd)
	ser.SetOnConnStop(OnConnectionLost)
	ser.SetAfterServe(AfterServe)
	ser.SetOnShutdown(OnShutdown)

	ser.AddRouter(0, &Api{})

	ser.Serve()
}

type Api struct {
	wss.BaseRouter
}

func (a *Api) Handle(req wss.IRequest) {
	log.Println(req.GetMsg())
	req.GetConnection().SendBuffMsg(req.GetMsgID(), req.GetData())
	if req.GetMsgID() == 0x02 {
		n := req.GetConnection().(*wss.Connection).WSServer.GetConnMgr().Len()
		req.GetConnection().SendBuffMsg(0x03, []byte(`{"conn":`+fmt.Sprintf("%d", n)+`}`))
	}
}

func OnConnectionAdd(conn wss.IConnection) {
	conn.SendBuffMsg(0x00, []byte(`Welcome to WSS`))
}

func OnConnectionLost(conn wss.IConnection) {
	log.Printf("[api] logout conn id %d\n", conn.GetConnID())
}

func AfterServe() {
	log.Println("[api] start serving")
	dataBuff := bytes.NewBuffer([]byte{})

	binary.Write(dataBuff, binary.LittleEndian, uint32(0x12345678))
	fmt.Println(hex.EncodeToString(dataBuff.Bytes()))

	dataBuff = bytes.NewBuffer([]byte{})
	binary.Write(dataBuff, binary.BigEndian, uint32(0x12345678))
	fmt.Println(hex.EncodeToString(dataBuff.Bytes()))
}

func OnShutdown() {
	log.Println("[api] start shutdown")

	<-time.After(1 * time.Second)
	log.Println("[api] finish shutdown")
}
