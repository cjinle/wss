package wss

type IRouter interface {
	PreHandle(request IRequest)
	Handle(request IRequest)
	PostHandle(request IRequest)
}

type BaseRouter struct{}

func (br *BaseRouter) PreHandle(request IRequest)  {}
func (br *BaseRouter) Handle(request IRequest)     {}
func (br *BaseRouter) PostHandle(request IRequest) {}
