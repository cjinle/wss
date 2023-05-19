package wss

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/gorilla/websocket"
)

type IConnection interface {
	Start()
	Stop()

	GetWSConnection() *websocket.Conn
	GetConnID() uint32
	RemoteAddr() net.Addr

	Send(data []byte) error
	SendMsg(msgID uint32, data []byte) error
	SendBuffMsg(msgID uint32, data []byte) error

	SetProperty(key string, value interface{})
	GetProperty(key string) (interface{}, error)
	RemoveProperty(key string)
}

type Connection struct {
	WSServer   IServer
	Conn       *websocket.Conn
	ConnID     uint32
	MsgHandler IMsgHandle

	ctx    context.Context
	cancel context.CancelFunc

	msgChan     chan []byte
	msgBuffChan chan []byte

	sync.RWMutex
	property     map[string]interface{}
	propertyLock sync.Mutex
	isClosed     bool
}

func NewConnection(server IServer, conn *websocket.Conn, connID uint32, msgHandler IMsgHandle) *Connection {
	c := &Connection{
		WSServer:    server,
		Conn:        conn,
		ConnID:      connID,
		isClosed:    false,
		MsgHandler:  msgHandler,
		msgChan:     make(chan []byte),
		msgBuffChan: make(chan []byte, DefaultMaxMsgChanLen),
		property:    make(map[string]interface{}),
	}

	c.WSServer.GetConnMgr().Add(c)

	return c
}

func (c *Connection) StartWriter() {
	for {
		select {
		case data := <-c.msgChan:
			if err := c.Conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
		case data, ok := <-c.msgBuffChan:
			if ok {
				if err := c.Conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				fmt.Println("msgBuffChan is Closed")
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Connection) StartReader() {
	defer c.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msgType, bs, err := c.Conn.ReadMessage()
			if err != nil {
				fmt.Println("read msg error ", err)
				return
			}

			fmt.Println("receive:", bs)

			if len(bs) < MESSAGE_HEAD_LEN {
				fmt.Println("pack len < 8", bs)
				continue
			}

			dp := NewDataPack()
			msg, err := dp.Unpack(bs[0:MESSAGE_HEAD_LEN])
			if err != nil {
				fmt.Println("unpack error ", msgType, err)
				continue
			}

			var data []byte
			if int(msg.GetDataLen()) <= len(bs)-MESSAGE_HEAD_LEN {
				data = bs[MESSAGE_HEAD_LEN : MESSAGE_HEAD_LEN+msg.GetDataLen()]
			}
			msg.SetData(data)

			req := Request{
				conn: c,
				msg:  msg,
			}

			if DefaultWorkerPoolSize > 0 {
				c.MsgHandler.SendMsgToTaskQueue(&req)
			} else {
				go c.MsgHandler.DoMsgHandler(&req)
			}
		}
	}
}

func (c *Connection) Start() {
	c.ctx, c.cancel = context.WithCancel(context.Background())

	go c.StartReader()
	go c.StartWriter()

	c.WSServer.CallOnConnStart(c)
}

func (c *Connection) Stop() {
	fmt.Println("Conn Stop ConnID", c.ConnID)
	c.Lock()
	defer c.Unlock()

	if c.isClosed {
		return
	}
	c.isClosed = true

	c.Conn.Close()
	c.cancel()

	c.WSServer.GetConnMgr().Remove(c)

	close(c.msgBuffChan)
}

func (c *Connection) GetWSConnection() *websocket.Conn {
	return c.Conn
}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) Send(data []byte) error {
	c.RLock()
	if c.isClosed {
		c.RUnlock()
		return errors.New("connection closed when send msg")
	}
	c.RUnlock()

	c.msgChan <- data

	return nil
}

func (c *Connection) SendMsg(msgID uint32, data []byte) error {
	c.RLock()
	if c.isClosed {
		c.RUnlock()
		return errors.New("connection closed when send msg")
	}
	c.RUnlock()

	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgID, data))
	if err != nil {
		fmt.Println("pack error msg id", msgID, err)
		return errors.New("pack error")
	}

	c.msgChan <- msg

	return nil
}

func (c *Connection) SendBuffMsg(msgID uint32, data []byte) error {
	c.RLock()
	if c.isClosed {
		c.RUnlock()
		return errors.New("connection closed when send msg")
	}
	c.RUnlock()

	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgID, data))
	if err != nil {
		fmt.Println("pack error msg id", msgID, err)
		return errors.New("pack error")
	}

	fmt.Println("send:", msg)

	c.msgBuffChan <- msg

	return nil
}

func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	}
	return nil, errors.New("no property found")
}

func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
