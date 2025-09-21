package SumoConfig

import (
	"encoding/xml"
	"logger"
	"typeDef"
)

type inductionLoop struct {
	ID          string `xml:"id,attr"`
	Lane        string `xml:"lane,attr"`
	Pos         string `xml:"pos,attr"`
	File        string `xml:"file,attr"`
	Period      string `xml:"period,attr"`
	FriendlyPos string `xml:"friendlyPos,attr,omitempty"`
	VTypes      string `xml:"vTypes,attr"`
}

type laneAreaDetector struct {
	ID          string `xml:"id,attr"`
	Lane        string `xml:"lane,attr"`
	Pos         string `xml:"pos,attr"`
	EndPos      string `xml:"endPos,attr"`
	File        string `xml:"file,attr"`
	Period      string `xml:"period,attr"`
	FriendlyPos string `xml:"friendlyPos,attr,omitempty"`
}

type detEntry struct {
	Lane string `xml:"lane,attr"`
	Pos  string `xml:"pos,attr"`
}

type detExit struct {
	Lane string `xml:"lane,attr"`
	Pos  string `xml:"pos,attr"`
}

type entryExitDetector struct {
	ID            string     `xml:"id,attr"`
	File          string     `xml:"file,attr"`
	Period        string     `xml:"period,attr"`
	OpenEntry     string     `xml:"openEntry,attr"`
	VTypes        string     `xml:"vTypes,attr"`
	DetectPersons string     `xml:"detectPersons,attr,omitempty"`
	DetEntry      []detEntry `xml:"detEntry"`
	DetExit       []detExit  `xml:"detExit"`
}

type DetectorsXml struct {
	typeDef.XmlParse
	XMLName           xml.Name            `xml:"additional"`
	InductionLoop     []inductionLoop     `xml:"inductionLoop"`
	EntryExitDetector []entryExitDetector `xml:"entryExitDetector"`
	LaneAreaDetector  []laneAreaDetector  `xml:"laneAreaDetector"`
}

func (This *DetectorsXml) ParseUrl(url string, obj interface{}) {
	This.XmlParse.ParseUrl(url, obj)
}

// 导出指定区域范围内的检测器            //导出文件路径
func (This *DetectorsXml) ExportByRect(exportfile string, scope Rect, net Net) {
	newDetectorsXml := &DetectorsXml{}
	newDetectorsXml.XMLName = This.XMLName
	for _, eeVal := range This.EntryExitDetector {
		eeDetector := &entryExitDetector{}
		*eeDetector = eeVal
		eeDetector.DetEntry = []detEntry{}
		eeDetector.DetExit = []detExit{}
		for _, eValue := range eeVal.DetEntry {
			if net.IsLane(eValue.Lane, scope) {
				eeDetector.DetEntry = append(eeDetector.DetEntry, eValue)
			}
		}
		for _, eValue := range eeVal.DetExit {
			if net.IsLane(eValue.Lane, scope) {
				eeDetector.DetExit = append(eeDetector.DetExit, eValue)
			}
		}
		if len(eeDetector.DetEntry) > 0 || len(eeDetector.DetExit) > 0 {
			newDetectorsXml.EntryExitDetector = append(newDetectorsXml.EntryExitDetector, *eeDetector)
		}
	}
	for _, lVal := range This.LaneAreaDetector {
		if net.IsLane(lVal.Lane, scope) {
			newDetectorsXml.LaneAreaDetector = append(newDetectorsXml.LaneAreaDetector, lVal)
		}
	}
	for _, lVal := range This.InductionLoop {
		if net.IsLane(lVal.Lane, scope) {
			newDetectorsXml.InductionLoop = append(newDetectorsXml.InductionLoop, lVal)
		}
	}
	logger.Info.Println("ExportByRect to:", exportfile)
	SaveDataToXML(exportfile, newDetectorsXml)
}
