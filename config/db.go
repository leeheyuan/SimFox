package config

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "simulation:simulationSaas@tcp(192.168.0.103:3306)/simulationtraffic?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	/*err = db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.SimulationProject{},
		&models.MapData{},
		&models.SimulationConfig{},
		&models.SimulationRun{},
		&models.SimulationResult{},
	)
	if err != nil {
		log.Fatal("failed to migrate:", err)
	}*/

	DB = db
}
