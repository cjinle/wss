package wss

import (
	"fmt"
	"log"
)

type IMsgHandle interface {
	DoMsgHandler(request IRequest)
	AddRouter(msgID uint32, router IRouter)
	StartWorkerPool()
	SendMsgToTaskQueue(request IRequest)
}

type MsgHandle struct {
	Apis           map[uint32]IRouter
	WorkerPoolSize uint32
	TaskQueue      []chan IRequest
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]IRouter),
		WorkerPoolSize: DefaultWorkerPoolSize,
		TaskQueue:      make([]chan IRequest, DefaultWorkerPoolSize),
	}
}

func (mh *MsgHandle) DoMsgHandler(request IRequest) {
	handler, ok := mh.Apis[request.GetMsgID()>>24] // 0x1000000 = 1
	if !ok {
		fmt.Println("api msgID", request.GetMsgID(), "is not FOUND!")
		return
	}

	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

func (mh *MsgHandle) AddRouter(msgID uint32, router IRouter) {
	if _, ok := mh.Apis[msgID]; ok {
		log.Fatalf("repeated api , msgID = %d\n", msgID)
	}
	mh.Apis[msgID] = router
}

func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan IRequest) {
	for {
		mh.DoMsgHandler(<-taskQueue)
	}
}

func (mh *MsgHandle) StartWorkerPool() {
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		mh.TaskQueue[i] = make(chan IRequest, DefaultMaxWorkerTaskLen)
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

func (mh *MsgHandle) SendMsgToTaskQueue(request IRequest) {
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	mh.TaskQueue[workerID] <- request
}
