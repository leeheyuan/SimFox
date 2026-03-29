package longConnection

import (
	"crypto/tls"
	"log"
	"logger"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

//var TimeOutMax time.Duration = 120

type __msg_send_ struct {
	t   int
	buf []byte
}

type IMsgHandler interface {
	OnRecvText(message []byte) error
	OnRecvBinary(message []byte) error
	OnClose(conn *Connection)
	OnTime(dt float64)
}

type Connection struct {
	ws             *websocket.Conn
	sendChan       chan __msg_send_
	address        string
	waitD          time.Duration
	heartbeatCount int32
	done           chan struct{}
	handler        IMsgHandler
}

func NewSever(ws *websocket.Conn, handler IMsgHandler) *Connection {
	conn := new(Connection)
	conn.sendChan = make(chan __msg_send_, 1024)
	conn.waitD = time.Second //time.Second
	conn.SetWs(ws)
	conn.handler = handler
	return conn
}

func NewClient(address string, handler IMsgHandler, iswait bool) *Connection {

	conn := new(Connection)
	conn.sendChan = make(chan __msg_send_, 1024)

	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	ws, _, err := dialer.Dial(address, nil)
	if err != nil {
		if iswait {
			for {
				time.Sleep(500 * time.Millisecond)
				logger.Info.Print("Connection ing")
				ws, _, err = websocket.DefaultDialer.Dial(address, nil)
				if err == nil {
					break
				}

			}
		} else {
			log.Fatal("连接错误:", err)
			return nil
		}
	}
	conn.ws = ws
	conn.address = address
	conn.waitD = time.Second //time.Second

	conn.done = make(chan struct{})
	conn.handler = handler
	return conn
}

func (This *Connection) readPump() {
	ws := This.ws
	defer func() {
		ws.Close()
		close(This.done)
	}()
	for {
		ws.SetReadDeadline(time.Now().Add(This.waitD * 180))
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			logger.Err.Printf("Read Message err:%s", err.Error())
			break
		}
		if messageType == websocket.BinaryMessage {
			err := This.handler.OnRecvBinary(message)
			if err != nil {
				logger.Err.Printf("Recv Binary Message err:%s", err.Error())
				break
			}
		}
		if messageType == websocket.TextMessage {
			err := This.handler.OnRecvText(message)
			if err != nil {
				logger.Err.Printf("Recv Text Message err:%s", err.Error())
				break
			}
		}
	}
}

func WebSocketPerClose(ws *websocket.Conn, closeCode int, message string) {
	closeMessage := websocket.FormatCloseMessage(closeCode, message)
	deadline := time.Now().Add(time.Second)
	ws.WriteControl(websocket.CloseMessage, closeMessage, deadline)
}

func WriteMessage(ws *websocket.Conn, message []byte) error {
	return ws.WriteMessage(websocket.TextMessage, message)
}

func (This *Connection) writePump() {
	ticker := time.NewTicker(10 * time.Second)
	ws := This.ws
	sendChan := This.sendChan
	defer func() {
		ticker.Stop()
		ws.Close()
		This.ws = nil
		//logger.Info.Println("defer connection exit")
		This.handler.OnClose(This)
	}()
	for {
		select {
		case <-This.done:
			//logger.Info.Println("connection done")
			return
		case message, ok := <-sendChan:
			if !ok {
				return
			}
			ws.SetWriteDeadline(time.Now().Add(This.waitD * 7))
			err := ws.WriteMessage(message.t, message.buf)
			if err != nil {
				logger.Err.Printf("Write Message err:%s", err.Error())
				return
			}
		case <-ticker.C:
			atomic.AddInt32(&This.heartbeatCount, 10)
			This.handler.OnTime(10)
		}
	}
}

func (This *Connection) Close(closeCode int, message string) {
	closeMessage := websocket.FormatCloseMessage(closeCode, message)
	deadline := time.Now().Add(time.Second)
	This.ws.WriteControl(websocket.CloseMessage, closeMessage, deadline)
	This.ws.Close()
}

func (This *Connection) ServeWs() {
	go This.writePump()
	This.readPump()
}

func (This *Connection) ClientWs() {
	go This.writePump()
	go This.readPump()
}

func (This *Connection) ImmediatelySend(Type int, Msg []byte) error {
	ws := This.ws
	ws.SetWriteDeadline(time.Now().Add(This.waitD * 7))
	err := ws.WriteMessage(int(Type), Msg)
	if err != nil {
		logger.Err.Printf("Immediately Write Message err:%s", err.Error())
		return err
	}
	return nil
}

func (This *Connection) ImmediatelyRead(waitTime int64) ([]byte, error) {
	ws := This.ws
	ws.SetReadDeadline(time.Now().Add(This.waitD * 32))
	_, message, err := ws.ReadMessage()
	return message, err
}

func (This *Connection) SendBinaryMsg(msg []byte) {
	select {
	case This.sendChan <- __msg_send_{
		t:   websocket.BinaryMessage,
		buf: msg,
	}:
	default:
	}
}

func (This *Connection) SendMsg(msg []byte) {
	select {
	case This.sendChan <- __msg_send_{
		t:   websocket.TextMessage,
		buf: msg,
	}:
	default:
	}
}

func (This *Connection) SetWs(ws *websocket.Conn) {
	This.ws = ws
	This.done = make(chan struct{})
}

func (This *Connection) Exist(ws *websocket.Conn) bool {
	return This.ws == ws
}

func (This *Connection) HaveIP(ws *websocket.Conn) bool {
	return This.ws.UnderlyingConn().RemoteAddr().String() == ws.UnderlyingConn().RemoteAddr().String()
}

func (This *Connection) IP() string {
	addr := This.ws.UnderlyingConn().RemoteAddr().String()
	return strings.Split(addr, ":")[0]
}
