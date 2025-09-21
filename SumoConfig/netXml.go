package SumoConfig

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"logger"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"typeDef"
	"utils"
)

type Rect struct {
	top    float64
	bottom float64
	left   float64
	right  float64
}

type Vector3d struct {
	x float64
	y float64
	z float64
}

func NewRect(left float64, right float64, bottom float64, top float64) Rect {

	return Rect{
		left:   left,
		right:  right,
		top:    top,
		bottom: bottom,
	}
}

func NewRectByUrl(url string) Rect {
	net := Net{}
	net.ParseUrl(url, &net)
	ConvBoundary := strings.Split(net.Location.ConvBoundary, ",")

	left, err := strconv.ParseFloat(ConvBoundary[0], 64)
	if err != nil {
		logger.Err.Print("err")
	}
	bottom, err := strconv.ParseFloat(ConvBoundary[1], 64)
	if err != nil {
		logger.Err.Print("err")
	}

	right, err := strconv.ParseFloat(ConvBoundary[2], 64)
	if err != nil {
		logger.Err.Print("err")
	}
	top, err := strconv.ParseFloat(ConvBoundary[3], 64)
	if err != nil {
		logger.Err.Print("err")
	}
	return Rect{top: top, bottom: bottom, left: left, right: right}
}

func (r *Rect) Top() float64 {
	return r.top
}

func (r *Rect) Bottom() float64 {
	return r.bottom
}

func (r *Rect) Right() float64 {
	return r.right
}

func (r *Rect) Left() float64 {
	return r.left
}

func (r *Rect) Size() (float64, float64) {

	return r.right - r.left, r.top - r.bottom
}

func (r *Rect) isIntersecting(other Rect) bool {
	if r.left > other.right || other.left > r.right {
		return false
	}
	if r.bottom > other.top || other.bottom > r.top {
		return false
	}
	return true
}

func (r *Rect) isInRange(x float64, y float64) bool {
	if y >= r.bottom && x >= r.left && x <= r.right && y <= r.top {
		return true
	}
	return false
}

// 定义 XML 对应的结构
type Net struct {
	typeDef.XmlParse
	XMLName              xml.Name `xml:"net"`
	Version              string   `xml:"version,attr"`
	XMLNSXsd             string   `xml:"xmlns:xsd,attr,omitempty"`
	XMLNSXsi             string   `xml:"xmlns:xsi,attr,omitempty"`
	JunctionCornerDetail string   `xml:"junctionCornerDetail,attr"`
	LimitTurnSpeed       string   `xml:"limitTurnSpeed,attr"`

	AreaNum     int               `xml:"areaNum,attr,omitempty"`
	Location    Location          `xml:"location"`
	Edges       []Edge            `xml:"edge"`
	Junctions   []Junction        `xml:"junction"`
	TlLogics    []TlLogic         `xml:"tlLogic"`
	Connections []Connection      `xml:"connection"`
	EdgeMap     map[string]int    `xml:"edgeMap,attr,omitempty"`
	LaneMap     map[string][2]int `xml:"laneMap,attr,omitempty"`
	JunctionMap map[string]int    `xml:"junctionMap,attr,omitempty"`
}

type NetMini struct {
	typeDef.XmlParse
	XMLName  xml.Name `xml:"net"`
	Version  string   `xml:"version,attr"`
	Location Location `xml:"location"`
	AreaNum  int      `xml:"areaNum,attr,omitempty"`
	EdgeLeng string   `xml:"edgeLeng,attr,omitempty"`
}

type Location struct {
	NetOffset     string `xml:"netOffset,attr"`
	ConvBoundary  string `xml:"convBoundary,attr"`
	OrigBoundary  string `xml:"origBoundary,attr"`
	ProjParameter string `xml:"projParameter,attr"`
}

type Edge struct {
	ID            string `xml:"id,attr"`
	Function      string `xml:"function,attr,omitempty"`
	CrossingEdges string `xml:"crossingEdges,attr,omitempty"`
	From          string `xml:"from,attr,omitempty"`
	To            string `xml:"to,attr,omitempty"`
	Priority      string `xml:"priority,attr,omitempty"`
	Shape         string `xml:"shape,attr,omitempty"`
	Lanes         []Lane `xml:"lane"`
}

type Lane struct {
	ID     string  `xml:"id,attr"`
	Index  int     `xml:"index,attr"`
	Speed  float64 `xml:"speed,attr"`
	Length float64 `xml:"length,attr"`
	Width  float64 `xml:"width,attr"`
	Shape  string  `xml:"shape,attr"`
}

func (This *Lane) getShape() Shape {
	var s Shape
	s = strings.Split(This.Shape, " ")
	return s
}

type Phase struct {
	Duration string `xml:"duration,attr"`
	State    string `xml:"state,attr"`
}

type TlLogic struct {
	ID        string  `xml:"id,attr"`
	ProgramID int     `xml:"programID,attr"`
	Offset    float64 `xml:"offset,attr"`
	Phase     []Phase `xml:"phase"`
	Type      string  `xml:"type,attr"`
}

type Request struct {
	Foes     string `xml:"foes,attr,omitempty"`
	Index    string `xml:"index,attr,omitempty"`
	Response string `xml:"response,attr,omitempty"`
	Cont     string `xml:"cont,attr,omitempty"`
}

type Junction struct {
	ID       string    `xml:"id,attr"`
	Type     string    `xml:"type,attr"`
	X        float64   `xml:"x,attr"`
	Y        float64   `xml:"y,attr"`
	Z        float64   `xml:"z,attr"`
	IncLanes string    `xml:"incLanes,attr"`
	IntLanes string    `xml:"intLanes,attr"`
	Shape    string    `xml:"shape,attr"`
	Request  []Request `xml:"request"`
}

type Connection struct {
	From       string  `xml:"from,attr"`
	To         string  `xml:"to,attr"`
	FromLane   string  `xml:"fromLane,attr"`
	ToLane     string  `xml:"toLane,attr"`
	KeepClear  int     `xml:"keepClear,attr,omitempty"`
	Via        string  `xml:"via,attr,omitempty"`
	Dir        string  `xml:"dir,attr"`
	State      string  `xml:"state,attr"`
	LinkIndex  string  `xml:"linkIndex,attr,omitempty"`
	Tl         string  `xml:"tl,attr,omitempty"`
	Visibility float64 `xml:"visibility,attr"`
}

func (This *Edge) GetPoint(from string, departLane string, departPos float32) (float32, float32) {
	return 0, 0
}

func (This *Net) Init() {
	This.EdgeMap = map[string]int{}
	This.LaneMap = map[string][2]int{}
	This.JunctionMap = map[string]int{}
	for index, v := range This.Edges {
		This.EdgeMap[v.ID] = index
		for lindex, lane := range v.Lanes {
			This.LaneMap[lane.ID] = [2]int{index, lindex}
		}
	}
	for index, v := range This.Junctions {
		This.JunctionMap[v.ID] = index
	}
}

func (This *Net) ParseFile(url string, obj interface{}) {
	This.XmlParse.ParseFile(url, obj)
	This.Init()
}
func (This *Net) ParseUrl(url string, obj interface{}) {
	This.XmlParse.ParseUrl(url, obj)
	This.Init()
}

func (This *Net) isHasEdge(id string) bool {
	_, ok := This.EdgeMap[id]
	if ok {
		return true
	}
	return false
}

func (This *Net) getLaneShape(id string) []*Shape {
	var ret []*Shape = make([]*Shape, 0)
	index, ok := This.EdgeMap[id]
	if ok {
		mEdge := This.Edges[index]
		for _, lane := range mEdge.Lanes {
			var s Shape
			s = strings.Split(lane.Shape, " ")
			ret = append(ret, &s)
		}
	}
	return ret
}

func (This *Net) getEdgeShape(id string) Shape {
	var ret Shape
	index, ok := This.EdgeMap[id]
	if ok {
		mEdge := This.Edges[index]
		ret = strings.Split(mEdge.Shape, " ")
	}
	return ret
}

func (This *Net) IsEdgeInRect(edge Edge, rect Rect) bool {
	positions := strings.Split(edge.Shape, " ")
	for _, v := range positions {
		position := strings.Split(v, ",")
		if len(position) >= 2 {
			x, errx := strconv.ParseFloat(position[0], 64)
			y, erry := strconv.ParseFloat(position[1], 64)
			if errx != nil {
				logger.Err.Fatalln("x parse err")
			}
			if erry != nil {
				logger.Err.Fatalln("y parse err")
			}
			if rect.isInRange(x, y) {
				return true
			}
		}
	}
	return false
}

func (This *Net) IsJunctionInRect(junction Junction, rect Rect) bool {
	if rect.isInRange(junction.X, junction.Y) {
		return true
	}
	return false
}

func (This *Net) IsLaneInRect(lane Lane, rect Rect) bool {
	positions := strings.Split(lane.Shape, " ")
	for _, v := range positions {
		position := strings.Split(v, ",")
		if len(position) >= 2 {
			x, errx := strconv.ParseFloat(position[0], 64)
			y, erry := strconv.ParseFloat(position[1], 64)
			if errx != nil {
				logger.Err.Fatalln("x parse err")
			}
			if erry != nil {
				logger.Err.Fatalln("y parse err")
			}
			if rect.isInRange(x, y) {
				return true
			}
		}
	}
	return false
}

func (This *Net) IsLane(ID string, rect Rect) bool {
	if ID == " " {
		logger.Warn.Println("a lane id is err")
		return false
	}
	value, ok := This.LaneMap[ID]
	if !ok {
		//logger.Warn.Println("a lane found", ID)
		return false
	}
	edge := This.Edges[value[0]]
	lane := edge.Lanes[value[1]]
	return This.IsLaneInRect(lane, rect)
}

func (This *Net) PointIsInEdge(ID string, rect *Rect, startPos float32) (bool, float64, float64) {

	if ID == " " {
		logger.Warn.Println("ID is null")
		return false, math.MaxFloat64, math.MaxFloat64
	}
	_, ok := This.EdgeMap[ID]
	if ok {
		shape := This.getEdgeShape(ID)
		x, y, _ := shape.GetPointOffset(float64(startPos))
		return rect.isInRange(x, y), x, y
	}
	return false, math.MaxFloat64, math.MaxFloat64

}

func (This *Net) GetAllEdgeLeng(rect *Rect) float64 {
	leng := float64(0)
	for _, v := range This.Edges {
		var shape Shape = strings.Split(v.Shape, " ")
		for i := 0; i < shape.Num()-1; i++ {
			x0, y0, z0 := shape.GetPoint(i)
			x1, y1, z1 := shape.GetPoint(i + 1)
			if x1 != math.MaxFloat64 {
				if rect.isInRange(x1, y1) && rect.isInRange(x0, y0) {
					leng += utils.CalculateDistance3D(x0, y0, z0, x1, y1, z1)
				}
			}
		}
	}
	return leng
}

func (This *Net) IsEdge(ID string, rect Rect) bool {
	if ID == " " {
		logger.Warn.Println("ID is null")
		return false
	}
	value, ok := This.EdgeMap[ID]
	if !ok {
		value = -1
		for i, val := range This.Edges {
			if val.ID == ID {
				value = i
				break
			}
		}
	}
	if value < 0 {
		//logger.Warn.Println("ID not found")
		return false
	}
	return This.IsEdgeInRect(This.Edges[value], rect)
}

func (This *Net) GetEdges(rect Rect) []string {
	var rets []string
	for _, val := range This.Edges {
		positions := strings.Split(val.Shape, " ")
		for _, v := range positions {
			position := strings.Split(v, ",")
			if len(position) >= 2 {
				x, errx := strconv.ParseFloat(position[0], 64)
				y, erry := strconv.ParseFloat(position[1], 64)
				if errx != nil {
					logger.Err.Fatalln("x parse err")
				}
				if erry != nil {
					logger.Err.Fatalln("y parse err")
				}
				if rect.isInRange(x, y) {
					rets = append(rets, val.ID)
					break
				}
			}
		}
	}
	return rets
}

func (This *Net) IsEdges(IDs []string) bool {
	for _, val := range This.Edges {
		if val.ID == IDs[0] {
			return true
		}
	}
	return false
}

func (This *Net) IsHaveEdge(EdgeID string) bool {
	_, ok := This.EdgeMap[EdgeID]
	return ok
}

func (This *Net) IsHaveLand(LandID string) bool {
	for _, e := range This.Edges {
		for _, v := range e.Lanes {
			if v.ID == LandID {
				return true
			}
		}
	}
	return false
}

// Flow 结构体表示 <flow> 元素
type Trip struct {
	ID     string `xml:"id,attr"`
	Depart string `xml:"depart,attr,omitempty"`
	From   string `xml:"from,attr"`
	To     string `xml:"to,attr"`
}

type Trips struct {
	XMLName xml.Name `xml:"trips"`
	Trip    []Trip   `xml:"trip"`
}

type Vehicles struct {
	typeDef.XmlParse
	XMLName xml.Name  `xml:"routes"`
	Vehicle []Vehicle `xml:"vehicle"`
}

type RouteNoID struct {
	XMLName xml.Name `xml:"route"`
	Edges   string   `xml:"edges,attr"` // 定义 edges 属性
}

type Vehicle struct {
	ID    string `xml:"id,attr"`
	Route Route  `xml:"route"`
}

type FlowEx struct {
	XMLName     xml.Name `xml:"flow"`
	ID          string   `xml:"id,attr"`
	Color       string   `xml:"color,attr,omitempty"`
	Begin       int      `xml:"begin,attr"`
	End         int      `xml:"end,attr"`
	Type        string   `xml:"type,attr"`
	From        string   `xml:"from,attr"`
	To          string   `xml:"to,attr"`
	Probability float64  `xml:"probability,attr"`
	DepartPos   float64  `xml:"departPos,attr,omitempty"`
	DepartLane  string   `xml:"departLane,attr,omitempty"`
	Route       Route    `xml:"route"`
}

type Roul struct {
	typeDef.XmlParse
	XMLName xml.Name `xml:"routes"`
	VType   []VType  `xml:"vType"`
	FlowEx  []FlowEx `xml:"flow"`
}

type IDArrayPing struct {
	typeDef.JsonParse
	Route []string `json:"route"`
	Type  []string `json:"type"`
	Edge  []string `json:"edge"`
	Tl    []string `json:"tl"`
	Lane  []string `json:"lane"`
}

type IDMapPing struct {
	Route map[string]int32
	Type  map[string]int32
	Edge  map[string]int32
	Tl    map[string]int32
	Lane  map[string]int32
}

var idArrayPing IDArrayPing
var idMapPing IDMapPing

func SaveDataToXML(filename string, data interface{}) error {
	// 创建 XML 文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	outputXML, err := xml.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling XML:", err)
		return nil
	}
	file.Write(outputXML)
	return nil
}

var flows Flows

func SplitRoulFile(roulfile string, savefile string) {
	var tabFlows []string
	file, err := os.Open(roulfile)
	if err != nil {
		fmt.Println("Error opening file:", err)
	}
	defer file.Close()
	// 创建 XML 解码器
	decoder := xml.NewDecoder(file)
	// 初始化一个 Net 对象来存储解析后的数据
	var roul Roul
	var vehicleTypes VehicleTypes

	// 开始解析 XML
	err = decoder.Decode(&roul)
	if err != nil {
		fmt.Println("error decoding XML:", err)
	}

	for _, v := range roul.FlowEx {
		v.Route.ID = v.ID
		if v.Route.Edges == "" {
			v.Route.Edges = v.From + " " + v.To
		}
		tabFlows = append(tabFlows, v.ID)
		vehicleTypes.Route = append(vehicleTypes.Route, v.Route)
		flows.Flows = append(flows.Flows, Flow{
			XMLName:     v.XMLName,
			ID:          v.ID,
			Color:       v.Color,
			Begin:       v.Begin,
			End:         v.End,
			Type:        v.Type,
			From:        v.From,
			To:          v.To,
			Probability: v.Probability,
			DepartPos:   v.DepartPos,
			DepartLane:  v.DepartLane,
			Route:       v.ID,
		})
	}

	idArrayPing.Route = tabFlows
	vehicleTypes.VType = append(vehicleTypes.VType, roul.VType...)
	vehicleTypes.XMLName = roul.XMLName

	SaveDataToXML(savefile+".roul.xml", vehicleTypes)
	flows.savefile = savefile + ".flow.xml"
	//SaveDataToXML(savefile+".flow.xml", flows)

}

func Save(r []Rect, neturl string) {
	//flows.SetAreaIndex()
	var net Net
	net.ParseUrl(neturl, &net)
	flows.SetAreaIndex(r, net)
	SaveDataToXML(flows.savefile, flows)
}

func SaveEx(r []*Rect, neturl string) {
	//flows.SetAreaIndex()
	var net Net
	net.ParseUrl(neturl, &net)
	flows.SetAreaIndexEx(r, net)
	SaveDataToXML(flows.savefile, flows)
}

func SaveIDMapPingEx(name string, roul Roul, net Net) {
	for _, v := range roul.FlowEx {
		idArrayPing.Route = append(idArrayPing.Route, v.ID)
	}

	for _, v := range roul.VType {
		idArrayPing.Type = append(idArrayPing.Type, v.ID)
	}

	logger.Info.Println("总共的道路:", len(net.Edges))
	for _, v := range net.Edges {
		idArrayPing.Edge = append(idArrayPing.Edge, v.ID)
		for _, Lane := range v.Lanes {
			idArrayPing.Lane = append(idArrayPing.Lane, Lane.ID)
		}
	}

	for _, v := range net.TlLogics {
		idArrayPing.Tl = append(idArrayPing.Tl, v.ID)
	}

	// 将 Person 结构体实例编码为 JSON 格式
	jsonData, err := json.Marshal(idArrayPing)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// 将 JSON 数据写入文件
	file, err := os.Create(name + ".json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing JSON to file:", err)
		return
	}
}

func SaveIDMapPingByFile(name string, roulurl string, neturl string) {
	var roul Roul
	var net Net
	roul.ParseFile(roulurl, &roul)
	net.ParseFile(neturl, &net)
	logger.Info.Println("SaveIDMapPing:", neturl)
	SaveIDMapPingEx(name, roul, net)
}

func SaveIDMapPing(name string, roulurl string, neturl string) {
	var roul Roul
	var net Net
	roul.ParseUrl(roulurl, &roul)
	net.ParseUrl(neturl, &net)
	logger.Info.Println("SaveIDMapPing:", neturl)
	SaveIDMapPingEx(name, roul, net)
}

func EdgesindexOf(arr []string, target string) int {
	for i, v := range arr {
		if v == target {
			return i
		}
	}
	return -1 // 如果没有找到，返回 -1
}

func calculateDistance3D(x1, y1, z1, x2, y2, z2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2)+math.Pow(y2-y1, 2)) + math.Pow(z2-z1, 2)
}

func Vector3disInRange(ponits []*Vector3d, r *Rect) bool {
	for _, v := range ponits {
		if r.isInRange(v.x, v.y) {
			return true
		}
	}
	return false
}

func getLaneShapesPoint(index int, shapes []*Shape) []*Vector3d {
	var rets []*Vector3d
	for _, shape := range shapes {
		if index < shape.Num() {
			x, y, z := shape.GetPoint(index)
			if x != math.MaxFloat64 && y != math.MaxFloat64 {
				rets = append(rets, &Vector3d{
					x: x,
					y: y,
					z: z,
				})
			}
		}
	}
	return rets
}

func SplitTest(route Route, net *Net, rect []*Rect, nets []*Net) []VehicleTypes {
	var subvehicleTypes []VehicleTypes = make([]VehicleTypes, len(rect))
	var selectIndexs []int = make([]int, len(rect))
	var newedgess []string = make([]string, len(rect))
	var isadds []bool = make([]bool, len(rect))
	x2, y2, z2 := math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
	Edges := strings.Split(route.Edges, " ")
	pos := 0.00
	for i := 0; i < len(selectIndexs); i++ {
		selectIndexs[i] = -1
		//isadds[i] = false
		newedgess[i] = ""
	}
	for _, e := range Edges {
		shapes := net.getLaneShape(e)
		shape := net.getEdgeShape(e)
		Num := shape.Num()

		for i := 0; i < len(selectIndexs); i++ {
			//selectIndexs[i] = -1
			isadds[i] = false
			//newedgess[i] = ""
		}

		for i := 0; i < 100000000; i++ {
			posints := getLaneShapesPoint(i, shapes)
			if i < Num && i >= 0 {
				x1, y1, z1 := shape.GetPoint(i)
				if x1 != math.MaxFloat64 && y1 != math.MaxFloat64 && x2 != math.MaxFloat64 && y2 != math.MaxFloat64 {
					pos += calculateDistance3D(x1, y1, z1, x2, y2, z2)
				}
				x2, y2, z2 = x1, y1, z1
			}
			if len(posints) == 0 {
				break
			}
			for i, v := range rect {
				if selectIndexs[i] >= 0 {
					if Vector3disInRange(posints, v) && nets[i].isHasEdge(e) {
						if !isadds[i] {
							if len(newedgess[i]) > 0 {
								newedgess[i] += " "
								newedgess[i] += e
							} else {
								newedgess[i] += e
							}
							isadds[i] = true
						}
					} else {
						subvehicleTypes[i].Route = append(subvehicleTypes[i].Route, Route{
							ID:    fmt.Sprintf("%s_%f", route.ID, pos),
							Edges: newedgess[i],
						})
						selectIndexs[i] = -1
						newedgess[i] = ""
					}
				}

				if selectIndexs[i] < 0 {
					if Vector3disInRange(posints, v) && nets[i].isHasEdge(e) {
						newedgess[i] += e
						selectIndexs[i] = i
						isadds[i] = true
					}
				}

			}
		}
	}

	for i, v := range selectIndexs {
		if v >= 0 {
			subvehicleTypes[i].Route = append(subvehicleTypes[i].Route, Route{
				ID:    fmt.Sprintf("%s_%f", route.ID, pos),
				Edges: newedgess[i],
			})
			selectIndexs[i] = -1
			newedgess[i] = ""
		}
	}

	for j, v := range nets {
		route := subvehicleTypes[j]
		for _, r := range route.Route {
			sEdges := strings.Split(r.Edges, " ")
			s := EdgesindexOf(Edges, sEdges[0])
			e := EdgesindexOf(Edges, sEdges[len(sEdges)-1])
			for i := s - 1; i >= 0; i-- {
				if v.isHasEdge(Edges[i]) {
					r.Edges = Edges[i] + " " + r.Edges
					continue
				}
				break
			}
			for i := e + 1; i < len(Edges); i++ {
				if v.isHasEdge(Edges[i]) {
					r.Edges = r.Edges + " " + Edges[i]
					continue
				}
				break
			}
		}

	}

	return subvehicleTypes
}

func LoadNet(netPath string) (*Rect, *Net) {
	net := Net{}
	net.ParseFile(netPath, &net)
	logger.Info.Println("netPath:", netPath)
	ConvBoundary := strings.Split(net.Location.ConvBoundary, ",")
	left, err := strconv.ParseFloat(ConvBoundary[0], 64)
	if err != nil {
		logger.Err.Print("err")
	}
	bottom, err := strconv.ParseFloat(ConvBoundary[1], 64)
	if err != nil {
		logger.Err.Print("err")
	}

	right, err := strconv.ParseFloat(ConvBoundary[2], 64)
	if err != nil {
		logger.Err.Print("err")
	}
	top, err := strconv.ParseFloat(ConvBoundary[3], 64)
	if err != nil {
		logger.Err.Print("err")
	}
	originalRect := Rect{top: top, bottom: bottom, left: left, right: right}
	return &originalRect, &net
}

func ExportLaneNetRoute(netPath string, roulfile string, exportfile string, index int) (Rect, Net) {
	logger.Info.Println("ExportLaneNetRoute:", netPath)
	var vehicleTypes VehicleTypes
	vehicleTypes.ParseFile(roulfile, &vehicleTypes)
	net := Net{}
	net.ParseFile(netPath, &net)
	var subvehicleTypes VehicleTypes
	subvehicleTypes.VType = vehicleTypes.VType
	logger.Info.Println("netPath:", netPath)
	ConvBoundary := strings.Split(net.Location.ConvBoundary, ",")
	left, err := strconv.ParseFloat(ConvBoundary[0], 64)
	if err != nil {
		logger.Err.Print("err")
	}
	bottom, err := strconv.ParseFloat(ConvBoundary[1], 64)
	if err != nil {
		logger.Err.Print("err")
	}

	right, err := strconv.ParseFloat(ConvBoundary[2], 64)
	if err != nil {
		logger.Err.Print("err")
	}
	top, err := strconv.ParseFloat(ConvBoundary[3], 64)
	if err != nil {
		logger.Err.Print("err")
	}
	originalRect := Rect{top: top, bottom: bottom, left: left, right: right}
	for _, v := range vehicleTypes.Route {
		newedges := ""
		Edges := strings.Split(v.Edges, " ")
		startIndex := -1
		startName := ""
		endName := ""
		for i, e := range Edges {
			if net.IsEdge(e, originalRect) {
				if len(newedges) > 0 {
					newedges += " "
					newedges += e
				} else {
					newedges += e
				}
				if startIndex == -1 {
					startName = e
					startIndex = i
				}
				endName = e
			} else if startIndex != -1 {
				startIndex = -1
				if len(newedges) > 0 {
					subvehicleTypes.Route = append(subvehicleTypes.Route, Route{
						ID:    fmt.Sprintf("%s_%s_%s", startName, v.ID, endName),
						Edges: newedges,
					})
				}
				newedges = ""
			}
		}
		if len(newedges) > 0 {
			subvehicleTypes.Route = append(subvehicleTypes.Route, Route{
				ID:    fmt.Sprintf("%s_%s_%s", startName, v.ID, endName),
				Edges: newedges,
			})
		}

	}
	SaveDataToXML(exportfile, subvehicleTypes)
	return originalRect, net
}

func SplitNetWork(netPath string, rect []Rect, roulfile string, savefile string) {
	var net Net
	net.ParseFile(netPath, &net)
	for i, v := range rect {
		var subnet Net
		subnet.Location.NetOffset = "0,0"
		subnet.Location.ConvBoundary = fmt.Sprintf("%f,%f,%f,%f", v.left, v.bottom, v.right, v.top)
		for i := 0; i < len(net.Junctions); i++ {
			junctions := net.Junctions[i]
			if v.isInRange(junctions.X, junctions.Y) {
				subnet.Junctions = append(subnet.Junctions, junctions)
			}
		}

		for i := 0; i < len(net.Edges); i++ {
			edge := net.Edges[i]
			if net.IsEdgeInRect(edge, v) {
				subnet.Edges = append(subnet.Edges, edge)
			}
		}

		SaveDataToXML(fmt.Sprintf("%s/%d.net.xml", savefile, i), subnet)
	}
}

// 来自sumo本身的路网配置
func SplitRoulFileEx(netPath string, roulfile string, savefile string) {

	file, err := os.Open(roulfile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	// 创建 XML 解码器
	decoder := xml.NewDecoder(file)
	// 初始化一个 Net 对象来存储解析后的数据
	var roul Roul
	var vehicleTypes VehicleTypes
	var flows Flows
	var trips Trips
	var vehicles Vehicles
	// 开始解析 XML
	err = decoder.Decode(&roul)
	if err != nil {
		fmt.Println("Error decoding XML:", err)
		return
	}

	for _, v := range roul.FlowEx {
		trips.Trip = append(trips.Trip, Trip{ID: v.ID, Depart: "0", From: v.From, To: v.To})
	}

	tripsfile := savefile + ".trips.xml"

	tripsroulfile := savefile + ".roul.trips.xml"

	SaveDataToXML(tripsfile, trips)

	var cmd *exec.Cmd

	cmd = exec.Command("duarouter", "--net-file", netPath, "--route-files", tripsfile, "--output-file", tripsroulfile, "--named-routes", "false", "--write-trips", "false")
	// 获取输出
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Start()
	if err == nil {
		//duarouter --net-file wutaishan.net.xml --route-files wts.trips.xml --output-file wts.trips.rou.xml --named-routes false --write-trips false
		vehicles.ParseFile(tripsroulfile, &vehicles)

		for _, v := range vehicles.Vehicle {
			v.Route.ID = v.ID
			vehicleTypes.Route = append(vehicleTypes.Route, v.Route)
		}
		for _, v := range roul.FlowEx {
			v.Route.ID = v.ID
			if v.Route.Edges == "" {
				v.Route.Edges = v.From + " " + v.To
				//trips.Trip = append(trips.Trip, Trip{ID: v.ID, Depart: "0", From: v.From, To: v.To})
			}
			//vehicleTypes.Route = append(vehicleTypes.Route, v.Route)
			flows.Flows = append(flows.Flows, Flow{
				XMLName:     v.XMLName,
				ID:          v.ID,
				Color:       v.Color,
				Begin:       v.Begin,
				End:         v.End,
				Type:        v.Type,
				From:        v.From,
				To:          v.To,
				Probability: v.Probability,
				DepartPos:   v.DepartPos,
				DepartLane:  v.DepartLane,
				Route:       v.ID,
			})
		}
		vehicleTypes.VType = append(vehicleTypes.VType, roul.VType...)
		vehicleTypes.XMLName = roul.XMLName

		SaveDataToXML(savefile+".roul.xml", vehicleTypes)
		SaveDataToXML(savefile+".flow.xml", flows)
	}

}
