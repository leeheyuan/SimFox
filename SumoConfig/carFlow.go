package SumoConfig

import (
	"logger"
	"math"
	"math/rand"
	"utils"
)

type CarFlow struct {
	flows          Flows
	vehicleTypes   VehicleTypes
	ids            []bool
	idFlowIndex    []int32
	departedsnum   int32
	isFlowdeparted []bool
	flowPoint      int
	AreaVehicleNum []int32
}

func (This *CarFlow) VehicleTypes() *VehicleTypes {
	return &This.vehicleTypes
}

func randNormal(mu, sigma float64) float64 {
	rand1 := rand.Float64()
	rand2 := rand.Float64()
	z := math.Sqrt(-2*math.Log(rand1)) * math.Cos(2*math.Pi*rand2)
	return mu + z*sigma
}

func (This *CarFlow) Departed(index int32) {
	This.isFlowdeparted[This.idFlowIndex[index]] = true
}

func (This *CarFlow) GetVehicleNum(index int) int32 {
	if len(This.AreaVehicleNum) <= index {
		additional := make([]int32, index-len(This.AreaVehicleNum)+1)
		This.AreaVehicleNum = append(This.AreaVehicleNum, additional...)
	}
	return This.AreaVehicleNum[index]
}

func (This *CarFlow) VehicleNum(index int, num int32) {
	if len(This.AreaVehicleNum) <= index {
		additional := make([]int32, index-len(This.AreaVehicleNum)+1)
		This.AreaVehicleNum = append(This.AreaVehicleNum, additional...)
	}
	This.AreaVehicleNum[index] = num
}

func (This *CarFlow) RemoveID(index int32) {
	//This.iDQueue.PushBack(index)
	This.ids[index] = false
	This.departedsnum--
	//This.isFlowdeparted[This.idFlowIndex[index]] = true
}

func (This *CarFlow) CheckAreaIndex(mRect []Rect) {

	for i := 0; i < len(This.flows.Flows); i++ {
		//This.flows.Flows[i].
		//This.flows.Flows[i].AreaIndex

	}

}

func (This *CarFlow) InitIDS(maxnumber int) {
	//maxnumber += 8000
	This.departedsnum = 0
	This.ids = make([]bool, maxnumber)
	This.idFlowIndex = make([]int32, maxnumber)
	This.AreaVehicleNum = make([]int32, 0)
	for i := 0; i < maxnumber; i++ {
		This.ids[i] = false
	}

	for i := 0; i < maxnumber; i++ {
		This.idFlowIndex[i] = -1
	}

}

func (This *CarFlow) InitCarFlow(roulurl string, flowurl string) {

	This.vehicleTypes.ParseUrl(roulurl, &This.vehicleTypes)
	This.flows.ParseUrl(flowurl, &This.flows)
	//	This.iDQueue = list.New()
	This.flowPoint = 0
	flowsleng := len(This.flows.Flows)
	This.isFlowdeparted = make([]bool, flowsleng)
	for i := 0; i < flowsleng; i++ {
		This.isFlowdeparted[i] = true
	}
}

func (This *CarFlow) GetID() (int32, bool) {
	for index, v := range This.ids {
		if !v {
			This.ids[index] = true
			This.departedsnum++
			return int32(index), true
		}
	}
	return int32(len(This.ids) + 1), false
}

func (This *CarFlow) RandomType() string {

	index := utils.RandomNum(0, len(This.vehicleTypes.VType)-1, nil)
	return This.vehicleTypes.VType[index].ID

}

func (This *CarFlow) Insert(myStep int64) ([]*Flow, []int32) {
	var ret []*Flow
	var Ids []int32
	count := len(This.flows.Flows)

	/*upperlimit := This.flowPoint + 4000
	for index := This.flowPoint; index < count && index < upperlimit; index++ {
		v := This.flows.Flows[index]
		if This.GetVehicleNum(v.AreaIndex) < 6000 && This.isFlowdeparted[index] {
			randomFloat := randNormal(0, 0.5)
			if randomFloat < v.Probability*0.5/1000 {
				ID, ok := This.GetID()
				if ok {
					ret = append(ret, &v)
					Ids = append(Ids, ID)
					This.idFlowIndex[ID] = int32(index)
					This.isFlowdeparted[index] = false
					This.flowPoint = (index + 1) % count
				} else {
					logger.Warn.Println("No ID available:", len(Ids), This.flowPoint)
					return ret, Ids
				}
			}
		}
	}*/

	for index := 0; index < count; index++ {
		v := This.flows.Flows[index]
		if This.isFlowdeparted[index] {
			//randomFloat := randNormal(0, 0.5)
			//if randomFloat < v.Probability*0.5/1000 {
			ID, ok := This.GetID()
			if ok {
				ret = append(ret, &v)
				Ids = append(Ids, ID)

				This.idFlowIndex[ID] = int32(index)
				This.isFlowdeparted[index] = false
			} else {
				logger.Warn.Println("No ID available:", len(Ids))
				return ret, Ids
			}
			//}
		}
	}
	logger.Warn.Println("Install Vehicle Num:", len(Ids), count)
	return ret, Ids
}
