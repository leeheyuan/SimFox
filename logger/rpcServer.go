package logger

type RPCServer struct {
}

func (_ *RPCServer) DisableMsg(req string, resp *string) (err error) {
	DisableMsg()
	return
}

func (_ *RPCServer) EnableMsg(req string, resp *string) (err error) {
	EnableMsg()
	return
}

func (_ *RPCServer) DisableInfo(req string, resp *string) (err error) {
	DisableInfo()
	return
}

func (_ *RPCServer) EnableInfo(req string, resp *string) (err error) {
	EnableInfo()
	return
}
