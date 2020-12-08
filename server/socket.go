package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

func NewKannaUnixSocketConnection(server *KannaServer, ID int, conn *net.UnixConn, handler MsgHandler) *KannaUnixSocketConnection {
	c := &KannaUnixSocketConnection{
		sKannaConnection: &sKannaConnection{
			Server:       server,
			ID:           ID,
			isClosed:     false,
			last:         time.Now().UnixNano(),
			msgHandler:   handler,
			ExitBuffChan: make(chan bool, 1),
			msgChan:      make(chan []byte),
			Props:        &sync.Map{},
		},

		Conn: conn,
	}

	c.Server.AddConn(c)
	return c
}

func (c *KannaUnixSocketConnection) SetAllowTelnet(to bool) {
	c.AllowTelnet = to
}

func (c *KannaUnixSocketConnection) SendMsg(data []byte) error {
	if c.isClosed {
		return fmt.Errorf("connection closed")
	}

	c.msgChan <- data

	return nil
}

func (c *KannaUnixSocketConnection) startWriter() {
	log.Println(c.ID, " Writer running")
	defer fmt.Println(c.ID, "[conn Writer exit!]")

	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				log.Println("Send Data Err:", err)
				return
			}

		case <-c.ExitBuffChan:
			break
		}

	}
}

func (c *KannaUnixSocketConnection) startReader() {
	log.Println(c.ID, "Reader running")

	defer c.Stop()
	for {
		var data []byte
		if _, err := io.ReadFull(c.Conn, data); err != nil {
			log.Println("read msg error", err)
			break
		}

		c.last = time.Now().UnixNano()
		req := Request{conn: c, msg: data}
		if c.msgHandler != nil {
			go c.msgHandler(&req)
		}
	}
}

func (c *KannaUnixSocketConnection) Start() {
	go c.startWriter()
	go c.startReader()
}

func (c *KannaUnixSocketConnection) Stop() {
	if c.isClosed {
		return
	}

	c.isClosed = true
	c.ExitBuffChan <- true

	// remove conn
	c.Server.RemoveConn(c)

	close(c.ExitBuffChan)
}

func (c *KannaUnixSocketConnection) GetID() int {
	return c.ID
}

func (c *KannaUnixSocketConnection) SetMsgHandler(h MsgHandler) {
	c.msgHandler = h
}

func (c *KannaUnixSocketConnection) GetProps() *sync.Map {
	return c.Props
}

func (c *KannaUnixSocketConnection) IsClosed() bool {
	return c.isClosed
}
