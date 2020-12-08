package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type KannaServer struct {
	ID            string
	OnConnStart   func(conn IKannaConnBehavior)
	OnConnEnd     func(conn IKannaConnBehavior)
	connections   map[int]IKannaConnBehavior
	SendCount     int
	NeedSendCount bool
	connLock      sync.RWMutex //读写连接的读写锁
}

var GServId = 0

func NewKananServer() *KannaServer {
	return &KannaServer{
		connections: make(map[int]IKannaConnBehavior),
	}
}

func (c *KannaServer) AddConn(conn IKannaConnBehavior) {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	c.connections[conn.GetID()] = conn
}

func (c *KannaServer) Broadcast(msg []byte) {
	for _, conn := range c.connections {
		conn.SendMsg(msg)
	}
}

func (c *KannaServer) SendMsg(id int, msg []byte) {
	c.connections[id].SendMsg(msg)
}

func (c *KannaServer) CloseAllConn() {
	for {
		if len(c.connections) < 1 {
			return
		}

		for _, conn := range c.connections {
			log.Println("Send close", conn)
			conn.Stop()
		}

		time.Sleep(time.Microsecond * 200)
	}
}

func (c *KannaServer) ExistsConn() bool {
	return len(c.connections) > 0
}

func (c *KannaServer) RemoveConn(conn IKannaConnBehavior) {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	delete(c.connections, conn.GetID())
}

func (c *KannaServer) Listen(addr string, port int, msgHandler MsgHandler) {
	if addr[0:7] == "socket:" && port == -1 {
		c.listenUnixSocket(addr[7:], msgHandler)
	} else {
		c.listenTCP(addr, port, msgHandler)
	}
}

func (c *KannaServer) listenTCP(addr string, port int, msgHandler MsgHandler) {
	go func() {
		addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", addr, port))
		log.Println("Listen TCP: ", addr)
		if err != nil {
			panic(err)
		}

		listener, err := net.ListenTCP("tcp4", addr)
		if err != nil {
			panic(err)
		}

		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				log.Println("tcp Accept err", err)
				continue
			}

			GServId++
			dealConn := NewKannaTcpConnection(c, GServId, conn, msgHandler)
			if c.OnConnStart != nil {
				c.OnConnStart(dealConn)
			}

			go dealConn.Start()
		}
	}()
}

func (c *KannaServer) listenUnixSocket(path string, msgHandler MsgHandler) {
	go func() {
		log.Println("Listen UnixSocket:", path)
		addr, err := net.ResolveUnixAddr("unix", path)
		if err != nil {
			log.Println("resolve unix addr err", err)
			return
		}

		listener, err := net.ListenUnix("unix", addr)
		if err != nil {
			log.Println("listen socket err", err)
			return
		}

		for {
			conn, err := listener.AcceptUnix()
			if err != nil {
				log.Println("unix Accept err", err)
				continue
			}

			GServId++
			dealConn := NewKannaUnixSocketConnection(c, GServId, conn, msgHandler)
			if c.OnConnStart != nil {
				c.OnConnStart(dealConn)
			}

			go dealConn.Start()
		}
	}()
}
