package typeDef

// 流的状态
type StreamerState int32

// 流的状态类型
const (
	Init  StreamerState = iota // 0
	Run                        // 1
	Idle                       // 2
	Close                      // 3
)

type ClientData struct {
	State StreamerState
	Id    string
	Name  string
}

type NodeEdgeLeng struct {
	Leng  float64 `json:"leng"`
	Index int32   `json:"index"`
}

func UnityExit() {

}

const INVALID_DOUBLE_VALUE = -1073741824.0
