package logger

import "testing"

func TestLogger(t *testing.T) {
	MsgRecv.Println("msg recv.")
	MsgSend.Println("msg send.")
	Err.Println("some error occur")
	Info.Println("what your name Time???")
}
