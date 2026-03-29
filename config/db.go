package config

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	//dsn := "simulation:simulationSaas@tcp(8.137.154.195:3306)/simulationtraffic?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := "root:sdahfjksahf897897hajkhdjsakdhahueu@tcp(8.137.154.195:3306)/simulationtraffic?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	 

	DB = db
}
