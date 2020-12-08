package utils

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

type WsConfig struct {
	Url     string
	Headers map[string][]string

	HeartbeatIntervalTime time.Duration
	HeartbeatData         func() []byte

	ProtoHandleFn func([]byte) error
	DecompressFn  func([]byte) ([]byte, error)
	OnError       func(error)
	OnConntected  func(conn *WsConn, isReconnect bool)

	IsAutoReconnect   bool
	ReconnectInterval time.Duration
	ReadDeadLineTime  time.Duration
}

type WsConn struct {
	conn *websocket.Conn
	WsConfig

	writeChan chan []byte
	pingChan  chan []byte
	pongChan  chan []byte
	closeChan chan []byte
	close     chan bool
	subs      [][]byte

	reconnLock *sync.Mutex
}

var WsDialer = &websocket.Dialer{
	Proxy:             http.ProxyFromEnvironment,
	HandshakeTimeout:  30 * time.Second,
	EnableCompression: true,
}

func (ws *WsConn) Init() *WsConn {
	if ws.HeartbeatIntervalTime == 0 {
		ws.ReadDeadLineTime = time.Minute
	} else {
		ws.ReadDeadLineTime = ws.HeartbeatIntervalTime * 2
	}

	if err := ws.connect(); err != nil {
		log.Panic(fmt.Errorf("[%s] %s", ws.Url, err.Error()))
	}

	ws.close = make(chan bool, 1)
	ws.pingChan = make(chan []byte, 10)
	ws.pongChan = make(chan []byte, 10)
	ws.closeChan = make(chan []byte, 10)
	ws.writeChan = make(chan []byte, 10)
	ws.reconnLock = new(sync.Mutex)

	go ws.writeRequest()
	go ws.receiveMessage()

	if ws.OnConntected != nil {
		ws.OnConntected(ws, false)
	}

	return ws
}

func (ws *WsConn) connect() error {
	wsConn, _, err := WsDialer.Dial(ws.Url, http.Header(ws.Headers))
	if err != nil {
		log.Printf("[%s] %s", ws.Url, err.Error())
		return err
	}

	wsConn.SetReadDeadline(time.Now().Add(ws.ReadDeadLineTime))
	ws.conn = wsConn

	log.Printf("[%s] connected", ws.Url)

	return nil
}

func (ws *WsConn) Reconnect() {
	log.Printf("[%s] Reconnecting..", ws.Url)

	ws.reconnLock.Lock()
	defer ws.reconnLock.Unlock()

	ws.conn.Close()

	var err error
	for {
		err = ws.connect()
		if err == nil {
			break
		}

		time.Sleep(ws.WsConfig.ReconnectInterval)
	}

	if ws.OnConntected != nil {
		ws.OnConntected(ws, true)
	}

	for _, sub := range ws.subs {
		log.Printf("[%s] re subscribe: ", ws.Url, string(sub))
		ws.SendMessage(sub)
	}
}

func (ws *WsConn) writeRequest() {
	var heartTimer *time.Timer
	var err error

	if ws.HeartbeatIntervalTime == 0 {
		heartTimer = time.NewTimer(time.Hour)
	} else {
		heartTimer = time.NewTimer(ws.HeartbeatIntervalTime)
	}

	for {
		select {
		case <-ws.close:
			log.Printf("[%s] closed", ws.Url)
			return
		case d := <-ws.writeChan:
			err = ws.conn.WriteMessage(websocket.TextMessage, d)
		case d := <-ws.pingChan:
			err = ws.conn.WriteMessage(websocket.PingMessage, d)
		case d := <-ws.closeChan:
			err = ws.conn.WriteMessage(websocket.CloseMessage, d)
		case <-heartTimer.C:
			if ws.HeartbeatIntervalTime > 0 {
				err = ws.conn.WriteMessage(websocket.TextMessage, ws.HeartbeatData())
				heartTimer.Reset(ws.HeartbeatIntervalTime)
			}
		}

		if err != nil {
			log.Printf("[%s]waitRequest exited", ws.Url)
		}
	}
}

func (ws *WsConn) Subscribe(subEvent interface{}) error {
	data, err := json.Marshal(subEvent)
	if err != nil {
		log.Printf("[%s] json encode error , %s", ws.Url, err)
		return err
	}

	ws.writeChan <- data
	ws.subs = append(ws.subs, data)
	return nil
}

func (ws *WsConn) SendMessage(msg []byte) {
	ws.writeChan <- msg
}

func (ws *WsConn) SendPingMessage(msg []byte) {
	ws.pingChan <- msg
}

func (ws *WsConn) SendPongMessage(msg []byte) {
	ws.pongChan <- msg
}

func (ws *WsConn) SendCloseMessage(msg []byte) {
	ws.closeChan <- msg
}

func (ws *WsConn) SendJsonMessage(m interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	ws.writeChan <- data
	return nil
}

func (ws *WsConn) receiveMessage() {
	ws.conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("[%s] websocket exiting [code=%d , text=%s]", ws.Url, code, text)
		//ws.CloseWs()
		return nil
	})

	ws.conn.SetPongHandler(func(pong string) error {
		log.Printf("[%s] received [pong] %s", ws.Url, pong)
		ws.conn.SetReadDeadline(time.Now().Add(ws.ReadDeadLineTime))
		return nil
	})

	ws.conn.SetPingHandler(func(ping string) error {
		log.Printf("[%s] received [ping] %s", ws.Url, ping)
		ws.conn.SetReadDeadline(time.Now().Add(ws.ReadDeadLineTime))
		return nil
	})

	for {
		select {
		case <-ws.close:
			log.Printf("[%s] close websocket, exiting receive message goroutine.", ws.Url)
			return
		default:
			t, msg, err := ws.conn.ReadMessage()
			if err != nil {
				log.Printf("[%s] %s", ws.Url, err.Error())
				if ws.IsAutoReconnect {
					log.Printf("[%s] unexpected closed, reconnecting..", ws.Url)
					ws.Reconnect()

					continue
				}

				if ws.OnError != nil {
					ws.OnError(err)
				}

				return
			}

			ws.conn.SetReadDeadline(time.Now().Add(ws.ReadDeadLineTime))
			switch t {
			case websocket.TextMessage:
				ws.ProtoHandleFn(msg)
			case websocket.BinaryMessage:
				if ws.DecompressFn == nil {
					ws.ProtoHandleFn(msg)
				} else {
					decoded, err := ws.DecompressFn(msg)

					if err != nil {
						log.Printf("[%s] decompress error %s", ws.Url, err.Error())
					} else {
						ws.ProtoHandleFn(decoded)
					}
				}
				//	case websocket.CloseMessage:
				//	ws.CloseWs()
			default:
				log.Printf("[%s] error websocket message type , content is :\n %s \n", ws.Url, string(msg))
			}
		}
	}
}

func (ws *WsConn) CloseWs() error {
	lastAR := ws.IsAutoReconnect
	ws.IsAutoReconnect = false // 如果手动关闭的话就不要触发重连了 不然无限套娃了草

	ws.close <- true // 关掉write和read

	close(ws.close)
	close(ws.writeChan)
	close(ws.closeChan)
	close(ws.pingChan)
	close(ws.pongChan)

	err := ws.conn.Close()
	if err != nil {
		log.Printf("[%s] close error %s", ws.Url, err.Error())
		return err
	}

	time.Sleep(time.Millisecond * 50) // 防止没有及时断开的情况
	ws.IsAutoReconnect = lastAR
	return nil
}
