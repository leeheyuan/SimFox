package SumoConfig

import (
	"logger"
	"math"
	"strconv"
	"strings"
	"utils"
)

type Shape []string

func (This *Shape) Num() int {
	return len(*This)
}

func (This *Shape) GetPointOffset(offset float64) (float64, float64, float64) {
	Num := This.Num()
	distance := 0.00
	if offset == 0 {
		return This.GetPoint(0)
	}
	for i := 0; i < Num-1; i++ {
		x1, y1, z1 := This.GetPoint(i)
		x2, y2, z2 := This.GetPoint(i + 1)
		if x2 == math.MaxFloat64 {
			break
		}
		distance += utils.CalculateDistance3D(x1, y1, z1, x2, y2, z2)
		if distance >= offset {
			return x2, y2, z2
		}
	}
	return math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
}

func (This *Shape) GetPoint(index int) (float64, float64, float64) {
	val := *This
	v := val[index]
	position := strings.Split(v, ",")
	if len(position) >= 2 {
		x, errx := strconv.ParseFloat(position[0], 64)
		y, erry := strconv.ParseFloat(position[1], 64)
		z := math.MaxFloat64
		if errx != nil {
			x = math.MaxFloat64
		}
		if erry != nil {
			logger.Err.Fatalln("y parse err")
			y = math.MaxFloat64
		}
		if len(position) >= 3 {
			z1, err := strconv.ParseFloat(position[2], 64)
			z = z1
			if err != nil {
				z = math.MaxFloat64
			}
		}
		return x, y, z
	}
	return math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
}
