package SumoConfig

import (
	"encoding/xml"
	"fmt"
	"logger"
	"typeDef"
)

type Route struct {
	XMLName xml.Name `xml:"route"`
	ID      string   `xml:"id,attr"`    // 定义 id 属性
	Edges   string   `xml:"edges,attr"` // 定义 edges 属性
}

type Routes struct {
	typeDef.XmlParse
	XMLName xml.Name `xml:"routes"`
	Route   []Route  `xml:"route"`
}

type VehicleTypes struct {
	typeDef.XmlParse
	XMLName xml.Name `xml:"routes"`
	VType   []VType  `xml:"vType"`
	Route   []Route  `xml:"route"`
	//typeMap map[string]int16
}

type VehicleMiniTypes struct {
	typeDef.XmlParse
	XMLName xml.Name `xml:"routes"`
	VType   []VType  `xml:"vType"`
}

func (This *VehicleTypes) ParseUrl(url string, obj interface{}) {
	logger.Info.Println("This.VType leng:", len(This.VType))
	This.XmlParse.ParseUrl(url, obj)
}

// Flow 结构体表示 <flow> 元素
type Flow struct {
	XMLName        xml.Name `xml:"flow"`
	ID             string   `xml:"id,attr"`
	Color          string   `xml:"color,attr,omitempty"`
	Begin          int      `xml:"begin,attr"`
	End            int      `xml:"end,attr"`
	AreaIndex      int      `xml:"areaIndex,attr"`
	Type           string   `xml:"type,attr"`
	From           string   `xml:"from,attr"`
	To             string   `xml:"to,attr"`
	Probability    float64  `xml:"probability,attr"`
	DepartPos      float64  `xml:"departPos,attr,omitempty"`
	DepartLane     string   `xml:"departLane,attr,omitempty"`
	Route          string   `xml:"route,attr"`
	DepartLocation string   `xml:"departLocation,attr,omitempty"`
}

type Flows struct {
	typeDef.XmlParse
	XMLName  xml.Name `xml:"routes"`
	Flows    []Flow   `xml:"flow"`
	savefile string
}

func (This *Flows) DownloadConfig(namefile string, neturl string, roulurl string) {

}

// seting flows of areaindex
func (This *Flows) SetAreaIndex(mRect []Rect, net Net) {
	for index, v := range This.Flows {
		ishave := false
		for i, r := range mRect {
			if v.From == "-3237" && i == 16 {
				logger.Info.Println("areaindex")
			}
			if net.IsEdge(v.From, r) {
				This.Flows[index].AreaIndex = i
				ishave = true
				break
			}
		}
		if !ishave {
			logger.Err.Println("a From ID no found:", v.From)
		}
	}

}

func (This *Flows) SetAreaIndexEx(mRect []*Rect, net Net) {
	for index, v := range This.Flows {
		ishave := false
		for i, r := range mRect {
			if v.From == "-3237" && i == 16 {
				logger.Info.Println("areaindex")
			}
			ok, x, y := net.PointIsInEdge(v.From, r, float32(v.DepartPos))
			if ok {
				This.Flows[index].AreaIndex = i
				This.Flows[index].DepartLocation = fmt.Sprintf("%f %f", x, y)
				ishave = true
				break
			}
		}
		if !ishave {
			logger.Err.Println("a From ID no found:", v.From)
		}
	}

}

func (This *Flows) ParseUrl(url string, obj interface{}) {
	This.XmlParse.ParseUrl(url, obj)
}
