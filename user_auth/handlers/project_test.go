package handlers

import (
	"config"
	"fmt"
	"log"
	"models"
	"os"
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
