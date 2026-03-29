package longConnection

import "logger"

type MessageBase struct {
}

func (This *MessageBase) OnClose(conn *Connection) {
	logger.Info.Println("Connection Close")
}

func (This *MessageBase) OnTime(dt float64) {

}

func (This *MessageBase) OnRecvBinary(message []byte) error {
	return nil
}

func (This *MessageBase) OnRecvText(message []byte) error {

	return nil
}
