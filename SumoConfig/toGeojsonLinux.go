//go:build linux
// +build linux

package SumoConfig

import (
	geojson "github.com/paulmach/go.geojson"
	"github.com/pebbe/go-proj-4/proj"
)

func ToGeojson(url string) []byte {
	var net Net
	net.ParseUrl(url, &net)
	pjSrc, _ := proj.NewProj("+proj=tmerc +lat_0=0 +lon_0=117 +k=1 +x_0=500000 +y_0=0 +ellps=GRS80 +units=m +no_defs")
	pjDst, _ := proj.NewProj("+proj=longlat +datum=WGS84 +no_defs")
	fc := geojson.NewFeatureCollection()
	for _, edge := range net.Edges {
		for _, lane := range edge.Lanes {
			shape := lane.getShape()
			var coords [][]float64
			for i := 0; i < shape.Num(); i++ {
				x, y, _ := shape.GetPoint(i)
				lon, lat, _ := proj.Transform2(pjSrc, pjDst, x, y)
				coords = append(coords, []float64{lon, lat})
			}
			line := geojson.NewLineStringFeature(coords)
			line.Properties["id"] = edge.ID
			line.Properties["from"] = edge.From
			line.Properties["to"] = edge.To
			fc.AddFeature(line)
		}
	}
	// 5. 写入 GeoJSON 文件
	geojsonData, _ := fc.MarshalJSON()
	return geojsonData
	//_ = ioutil.WriteFile("network.geojson", geojsonData, 0644)
}
