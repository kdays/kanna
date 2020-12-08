package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

func NewKannaTcpConnection(server *KannaServer, ID int, conn *net.TCPConn, handler MsgHandler) *KannaTCPConnection {
	c := &KannaTCPConnection{
		sKannaConnection: &sKannaConnection{
			Server:       server,
			ID:           ID,
			isClosed:     false,
			last:         time.Now().UnixNano(),
			msgHandler:   handler,
			ExitBuffChan: make(chan bool, 1),
			msgChan:      make(chan []byte),
			AllowTelnet:  false,
			Props:        &sync.Map{},
		},

		Conn: conn,
	}

	c.Server.AddConn(c)
	return c
}

func (c *KannaTCPConnection) GetID() int {
	return c.ID
}

func (c *KannaTCPConnection) SetAllowTelnet(to bool) {
	c.AllowTelnet = to
}

func (c *KannaTCPConnection) SendMsg(data []byte) error {
	if c.isClosed {
		return fmt.Errorf("connection closed")
	}

	///c.Server.SendCount++
	if c.Server.NeedSendCount {
		c.Server.SendCount++
	}

	select {
	case <-c.ExitBuffChan:
		return fmt.Errorf("conn is closed")
	default:
		c.msgChan <- data
	}

	return nil
}

func (c *KannaTCPConnection) startWriter() {
	log.Println(c.ID, "Writer running")

	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				log.Println("Send Data Err:", err)
				return
			}

		case <-c.ExitBuffChan:
			return
		}

	}
}

const PackHeadLen = 4

func (c *KannaTCPConnection) startReader() {
	log.Println(c.ID, "Reader running")
	log.Println(c.Conn)

	defer c.Stop()
	for {
		if c.isClosed == true {
			break
		}

		if c.AllowTelnet {
			data := make([]byte, 128)
			size, err := c.Conn.Read(data)
			if err != nil {
				log.Println("read msg error", err)
				break
			}

			if size > 0 {
				req := Request{conn: c, msg: data}
				if c.msgHandler != nil {
					go c.msgHandler(&req)
				}
			}
		} else {
			headData := make([]byte, PackHeadLen)
			if _, err := io.ReadFull(c.Conn, headData); err != nil {
				fmt.Println("read msg head error ", err)
				return
			}

			dataBuff := bytes.NewReader(headData)
			var msgLen uint32
			if err := binary.Read(dataBuff, binary.BigEndian, &msgLen); err != nil {
				fmt.Println("read msg data error - binary read", err)
				return
			}

			if msgLen > 1000 {
				fmt.Println("message too large", msgLen)
			}

			if msgLen > 0 && msgLen < 1000 {
				data := make([]byte, msgLen)
				if _, err := io.ReadFull(c.Conn, data); err != nil {
					fmt.Println("read msg data error ", err)
					return
				}

				req := Request{conn: c, msg: data}

				if c.msgHandler != nil {
					go c.msgHandler(&req)
				}
			}
		}
	}
}

func (c *KannaTCPConnection) Start() {
	go c.startWriter()
	go c.startReader()
}

func (c *KannaTCPConnection) Stop() {
	if c.isClosed == true {
		return
	}

	log.Println(c.ID, " will quit")
	c.isClosed = true

	c.ExitBuffChan <- true
	c.Conn.Close()

	if c.Server.OnConnEnd != nil {
		c.Server.OnConnEnd(c)
	}

	c.Server.RemoveConn(c)
	close(c.ExitBuffChan)
}

func (c *KannaTCPConnection) SetMsgHandler(h MsgHandler) {
	c.msgHandler = h
}

func (c *KannaTCPConnection) GetProps() *sync.Map {
	return c.Props
}

func (c *KannaTCPConnection) IsClosed() bool {
	return c.isClosed
}
