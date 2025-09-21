package SumoConfig

import "encoding/xml"

type VType struct {
	XMLName         xml.Name `xml:"vType"`
	ID              string   `xml:"id,attr"`
	Accel           float64  `xml:"accel,attr"`
	Decel           float64  `xml:"decel,attr"`
	Sigma           float64  `xml:"sigma,attr"`
	MinGap          float64  `xml:"minGap,attr"`
	Tau             float64  `xml:"tau,attr"`
	Length          float64  `xml:"length,attr"`
	Width           float64  `xml:"width,attr"`
	MaxSpeed        float64  `xml:"maxSpeed,attr"`
	LaneChangeModel string   `xml:"laneChangeModel,attr"`
	CarFollowModel  string   `xml:"carFollowModel,attr"`
	VClass          string   `xml:"vClass,attr"`
	SpeedFactor     float64  `xml:"speedFactor,attr"`
	SpeedDev        float64  `xml:"speedDev,attr"`
}
