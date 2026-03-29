package handlers

import (
	"bytes"
	"config"
	"fmt"
	"io/ioutil"
	"log"
	"models"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"port"
	"syscall"
	"testing"
	"utils"
)

func TestMigrate(t *testing.T) {
	config.InitDB()
	db := config.DB
	err := db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.SimulationProject{},
		&models.MapData{},
		&models.SimulationConfig{},
		&models.SimulationTask{},
		&models.SimulationResult{},
		&models.Simulation{},
	)
	if err != nil {
		log.Fatal("failed to migrate:", err)
	}
}

func TestSumoTask(t *testing.T) {
	utils.LunchSumoSimulation("../../sim_server", "E:/SimFox/user_auth/SimulationConfig/liheyuan/terst4/sumo.sumocfg")
}

func TestPort(t *testing.T) {
	portManager := port.NewPortManager(8000, 9000)

	//py sim_server.py --config=D:/item/SimFox/user_auth/SimulationConfig/liheyuan/test3/sumo.sumocfg --projectId=1 --ports=8010,8001
	portManager.StartProcess("py", 2, "../../sim_server/sim_server.py", fmt.Sprintf("--config=%s", "D:/item/SimFox/user_auth/SimulationConfig/liheyuan/test3/sumo.sumocfg"), "--projectId=1")
	//portManager.StartProcess("py", 2, "../../sim_server/sim_server.py", fmt.Sprintf("--config=%s", "E:/SimFox/user_auth/SimulationConfig/liheyuan/terst4/sumo.sumocfg"), "--projectId=2")
	// 创建一个通道用于接收信号
	sigCh := make(chan os.Signal, 1)
	// 将指定的信号发送到通道
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("等待信号输入 (Ctrl+C 来发送 SIGINT 或者使用 kill 命令发送 SIGTERM)")
	// 阻塞程序，直到收到指定的信号
	sig := <-sigCh
	fmt.Println("接收到信号:", sig)
}


func TestOverpassRequest(t *testing.T) {
    query := `
[out:json];
way["highway"](30.535,104.064,30.538,104.069);
out body;
>;
out skel qt;
`

    resp, err := http.Post(
        "https://overpass.kumi.systems/api/interpreter",
        "text/plain",
        bytes.NewBufferString(query),
    )

    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Println(string(body))
}


// Overpass 查询（XML格式，适合 SUMO）
const query = `
[out:xml];
(
  way["highway"](30.535,104.064,30.538,104.069);
);
(._;>;);
out body;
`

func TestOverpassRequestEx(t *testing.T) {
	osmFile := "map.osm"
	netFile := "net.xml"

	// 1️⃣ 请求 Overpass API
	body, err := fetchOSM(query)
	if err != nil {
		panic(err)
	}

	// 2️⃣ 保存为 .osm 文件
	err = saveToFile(osmFile, body)
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ OSM saved:", osmFile)

	// 3️⃣ 调用 netconvert
	err = runNetconvert(osmFile, netFile)
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ SUMO net generated:", netFile)
}

func TestGenerateRoute(t *testing.T) {
    netFile := "test.net.xml"

	routesFile, err := GenerateRoutes(netFile)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ 生成完成:", routesFile)
}

// 请求 Overpass（带简单容灾）
func fetchOSM(query string) ([]byte, error) {
	servers := []string{
		"https://overpass.kumi.systems/api/interpreter",
		"https://overpass.openstreetmap.fr/api/interpreter",
		"https://overpass-api.de/api/interpreter",
	}

	for _, url := range servers {
		fmt.Println("请求:", url)

		resp, err := http.Post(url, "text/plain", bytes.NewBufferString(query))
		if err != nil {
			fmt.Println("失败:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Println("状态异常:", resp.Status)
			continue
		}

		return ioutil.ReadAll(resp.Body)
	}

	return nil, fmt.Errorf("所有 Overpass 节点失败")
}

// 保存文件
func saveToFile(filename string, data []byte) error {
	return ioutil.WriteFile(filename, data, 0644)
}

// 调用 netconvert
func runNetconvert(osmFile, netFile string) error {
	cmd := exec.Command(
		"netconvert",
		"--osm-files", osmFile,
		"--output-file", netFile,
		"--geometry.remove",
		"--ramps.guess",
		"--junctions.join",
		"--tls.guess-signals",
		"--tls.discard-simple",
		"--tls.join",
		"--remove-edges.isolated",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}



// 生成随机 trips + routes
func GenerateRoutes(netFile string) (string, error) {
	tripsFile := "trips.trips.xml"
	routesFile := "routes.rou.xml"

	// 1️⃣ 调用 randomTrips.py
	err := runRandomTrips(netFile, tripsFile)
	if err != nil {
		return "", err
	}

	// 2️⃣ 调用 duarouter
	err = runDuaRouter(netFile, tripsFile, routesFile)
	if err != nil {
		return "", err
	}

	return routesFile, nil
}

// 调用 randomTrips.py
func runRandomTrips(netFile, tripsFile string) error {
	
	cmd := exec.Command( 
		"C:\\Users\\liheyuan\\AppData\\Local\\Programs\\Python\\Python314\\python.exe", "C:\\Program Files (x86)\\Eclipse\\Sumo\\tools\\randomTrips.py",
		"-n", netFile,
		"-o", tripsFile,
		"--begin", "0",
		"--end", "3600",
		"--period", "1", // 每秒一辆车
		"--fringe-factor","10",
        "--min-distance", "300",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("🚗 生成 trips...")
	return cmd.Run()
}

// 调用 duarouter
func runDuaRouter(netFile, tripsFile, routesFile string) error {
	cmd := exec.Command(
		"duarouter",
		"-n", netFile,
		"-t", tripsFile,
		"-o", routesFile,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("🧠 计算路径...")
	return cmd.Run()
}