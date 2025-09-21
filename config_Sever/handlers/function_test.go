package handlers

import (
	"config"
	"log"
	"models"
	"testing"
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
