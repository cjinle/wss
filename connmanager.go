package wss

import (
	"errors"
	"fmt"
	"sync"
)

type IConnManager interface {
	Add(conn IConnection)
	Remove(conn IConnection)
	Get(connID uint32) (IConnection, error)
	Len() int
	ClearConn()
}

type ConnManager struct {
	connections map[uint32]IConnection
	connLock    sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]IConnection),
	}
}

func (cm *ConnManager) Add(conn IConnection) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	cm.connections[conn.GetConnID()] = conn
	fmt.Println("[ConnManager] current conn len", cm.Len())
}

func (cm *ConnManager) Remove(conn IConnection) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	delete(cm.connections, conn.GetConnID())
	fmt.Println("[ConnManager] Remove", conn.GetConnID(), ", current conn len", cm.Len())
}

func (cm *ConnManager) Get(connID uint32) (IConnection, error) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	if conn, ok := cm.connections[connID]; ok {
		return conn, nil
	}
	return nil, errors.New("connection not found")
}

func (cm *ConnManager) Len() int {
	return len(cm.connections)
}

func (cm *ConnManager) ClearConn() {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	for connID, conn := range cm.connections {
		conn.Stop()
		delete(cm.connections, connID)
	}
	fmt.Println("[ConnManager] Clear All Connections, current conn len", cm.Len())
}
