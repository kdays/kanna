package server

import (
	"net"
	"sync"
)

type MsgHandler = func(d *Request)

// 读取逻辑在MsgHandler里
type IKannaConnBehavior interface {
	Start()
	Stop()
	IsClosed() bool
	GetID() int
	SendMsg(data []byte) error
	SetMsgHandler(h MsgHandler)
	GetProps() *sync.Map
	SetAllowTelnet(to bool)
}

type IRequest interface {
	GetConnection() IKannaConnBehavior
	GetData() []byte
}

type Request struct {
	conn IKannaConnBehavior
	msg  []byte
}

func (r *Request) GetConnection() IKannaConnBehavior {
	return r.conn
}

func (r *Request) GetData() []byte {
	return r.msg
}

type sKannaConnection struct {
	Server       *KannaServer
	ID           int
	last         int64
	msgHandler   MsgHandler
	isClosed     bool
	msgChan      chan []byte
	ExitBuffChan chan bool
	Props        *sync.Map
	AllowTelnet  bool
}

type KannaTCPConnection struct {
	*sKannaConnection
	Conn *net.TCPConn
}

type KannaUnixSocketConnection struct {
	*sKannaConnection
	Conn *net.UnixConn
}
