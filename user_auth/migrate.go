package main

import (
	"config"
	"models"
)

func autoMigrate() error {
	return config.DB.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.SimulationProject{},
		&models.MapData{},
		&models.SimulationConfig{},
		&models.SimulationTask{},
		&models.SimulationResult{},
		&models.Simulation{},
	)
}
