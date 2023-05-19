package wss

const (
	DefaultMaxPacketSize    = 4096  // 都需数据包的最大值
	DefaultMaxConn          = 12000 // 当前服务器主机允许的最大链接个数
	DefaultWorkerPoolSize   = 10    // 业务工作Worker池的数量
	DefaultMaxWorkerTaskLen = 1024  // 业务工作Worker对应负责的任务队列最大任务存储数量
	DefaultMaxMsgChanLen    = 1024  // SendBuffMsg发送消息的缓冲最大长度
)
