package SumoConfig

import (
	"encoding/xml"
)

type SumoConfiguration struct {
	XMLName           xml.Name `xml:"sumoConfiguration"`
	XmlnsXsi          string   `xml:"xmlns:xsi,attr"`
	XsiSchemaLocation string   `xml:"xsi:noNamespaceSchemaLocation,attr"`

	Input      *Input      `xml:"input"`
	Time       *Time       `xml:"time"`
	Processing *Processing `xml:"processing,omitempty"`
	Routing    *Routing    `xml:"routing,omitempty"`
	Report     *Report     `xml:"report,omitempty"`
	GuiOnly    *GuiOnly    `xml:"gui_only,omitempty"`
}

type Input struct {
	NetFile         StringAttr  `xml:"net-file,omitempty"`
	RouteFiles      StringAttr  `xml:"route-files,omitempty"`
	AdditionalFiles *StringAttr `xml:"additional-files,omitempty"`
}

type Time struct {
	End        IntAttr   `xml:"end"`
	StepLength FloatAttr `xml:"step-length"`
}

type Processing struct {
	IgnoreRouteErrors BoolAttr `xml:"ignore-route-errors"`
	TimeToTeleport    IntAttr  `xml:"time-to-teleport"`
}

type Routing struct {
	AdaptationSteps IntAttr `xml:"device.rerouting.adaptation-steps"`
}

type Report struct {
	Verbose            BoolAttr   `xml:"verbose"`
	XMLValidation      StringAttr `xml:"xml-validation"`
	DurationStatistics BoolAttr   `xml:"duration-log.statistics"`
	NoStepLog          BoolAttr   `xml:"no-step-log"`
}

type GuiOnly struct {
	GUISettingsFile StringAttr `xml:"gui-settings-file"`
}

// 通用属性封装
type StringAttr struct {
	Value string `xml:"value,attr,omitempty"`
}

type IntAttr struct {
	Value int `xml:"value,attr,omitempty"`
}

type FloatAttr struct {
	Value float64 `xml:"value,attr,omitempty"`
}

type BoolAttr struct {
	Value bool `xml:"value,attr,omitempty"`
}
