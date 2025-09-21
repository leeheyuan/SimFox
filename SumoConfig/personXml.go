package SumoConfig

import "encoding/xml"

type PersonFlow struct {
	XMLName        xml.Name `xml:"flow"`
	ID             string   `xml:"id,attr"`
	Begin          int      `xml:"begin,attr"`
	End            int      `xml:"end,attr"`
	AreaIndex      int      `xml:"areaIndex,attr"`
	Type           string   `xml:"type,attr"`
	Probability    float64  `xml:"probability,attr"`
	DepartPos      float64  `xml:"departPos,attr,omitempty"`
	DepartLane     string   `xml:"departLane,attr,omitempty"`
	Route          string   `xml:"route,attr"`
	DepartLocation string   `xml:"departLocation,attr,omitempty"`
}
