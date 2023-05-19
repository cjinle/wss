package wss

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
)

type IServer interface {
	Start()
	Stop()
	Serve()
	AddRouter(msgID uint32, router IRouter)
	GetConnMgr() IConnManager
	SetOnConnStart(func(IConnection))
	SetOnConnStop(func(IConnection))
	SetOnShutdown(func())
	SetAfterServe(func())
	SetMaxConn(num int)
	CallOnConnStart(conn IConnection)
	CallOnConnStop(conn IConnection)
}

type Server struct {
	Name        string
	IP          string
	Port        int
	Path        string
	MaxConn     int
	msgHandler  IMsgHandle
	ConnMgr     IConnManager
	OnConnStart func(conn IConnection)
	OnConnStop  func(conn IConnection)
	OnShutdown  func()
	AfterServe  func()
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewServer(addr, path, name string) IServer {
	arr := strings.Split(addr, ":")
	if len(arr) < 2 {
		fmt.Println("[Server] addr error")
	}
	port, err := strconv.Atoi(arr[1])
	if err != nil {
		fmt.Println("[Server] port error", err)
	}

	s := &Server{
		Name:       name,
		IP:         arr[0],
		Port:       port,
		Path:       path,
		MaxConn:    DefaultMaxConn,
		msgHandler: NewMsgHandle(),
		ConnMgr:    NewConnManager(),
	}
	return s
}

func (s *Server) Start() {
	fmt.Printf("[Server] start %s on %s:%d\n", s.Name, s.IP, s.Port)

	go func() {
		s.msgHandler.StartWorkerPool()
		var cid uint32 = 0
		http.HandleFunc(s.Path, func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Print("upgrade:", err)
				return
			}

			if s.ConnMgr.Len() > s.MaxConn {
				conn.Close()
				fmt.Println("[Server] max conn count", s.MaxConn, conn.RemoteAddr().String())
				return
			}
			fmt.Println("[Server] conn remote addr", conn.RemoteAddr().String())

			dealConn := NewConnection(s, conn, cid, s.msgHandler)
			cid++
			go dealConn.Start()
		})

		log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", s.IP, s.Port), nil))
		fmt.Println("[Server] start listen succ")
	}()
}

func (s *Server) Stop() {
	fmt.Printf("[Server] %s stoped.", s.Name)
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	s.Start()

	if s.AfterServe != nil {
		s.AfterServe()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	fmt.Printf("[Server] shutdown %v\n", <-c)
	s.CallOnShutdown()
}

func (s *Server) AddRouter(msgID uint32, router IRouter) {
	s.msgHandler.AddRouter(msgID, router)
}

func (s *Server) GetConnMgr() IConnManager {
	return s.ConnMgr
}

func (s *Server) SetOnConnStart(hookFunc func(IConnection)) {
	s.OnConnStart = hookFunc
}

func (s *Server) SetOnConnStop(hookFunc func(IConnection)) {
	s.OnConnStop = hookFunc
}

func (s *Server) SetOnShutdown(hookFunc func()) {
	s.OnShutdown = hookFunc
}

func (s *Server) SetAfterServe(hookFunc func()) {
	s.AfterServe = hookFunc
}

func (s *Server) SetMaxConn(num int) {
	if num > 0 {
		s.MaxConn = num
	}
}

func (s *Server) CallOnConnStart(conn IConnection) {
	if s.OnConnStart != nil {
		s.OnConnStart(conn)
	}
}

func (s *Server) CallOnConnStop(conn IConnection) {
	if s.OnConnStop != nil {
		s.OnConnStop(conn)
	}
}

func (s *Server) CallOnShutdown() {
	if s.OnShutdown != nil {
		s.OnShutdown()
	}
}
