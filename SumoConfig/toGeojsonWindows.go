//go:build windows
// +build windows

package SumoConfig

import (
	geojson "github.com/paulmach/go.geojson"
)

func ToGeojson(url string) []byte {

	var net Net
	net.ParseUrl(url, &net)

	// 2. 初始化坐标转换器（以中国高德投影为例）
	//pjSrc, _ := proj.NewProj("+proj=tmerc +lat_0=0 +lon_0=117 +k=1 +x_0=500000 +y_0=0 +ellps=GRS80 +units=m +no_defs") // SUMO 原始坐标
	//pjDst, _ := proj.NewProj("+proj=longlat +datum=WGS84 +no_defs")                                                    // 目标坐标（经纬度）
	// 3. 构建 Node 坐标索引

	// 4. 构建 GeoJSON Feature 集
	fc := geojson.NewFeatureCollection()

	for _, edge := range net.Edges {
		for _, lane := range edge.Lanes {
			// shape 格式是 "x1,y1 x2,y2 ..."

			shape := lane.getShape()
			var coords [][]float64

			for i := 0; i < shape.Num(); i++ {
				lon, lat, _ := shape.GetPoint(i)
				//lon, lat, _ := x, y //proj.Transform2(pjSrc, pjDst, x, y)
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
