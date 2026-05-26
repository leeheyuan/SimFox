package main

import (
	"config"
	"models"
)

func autoMigrate() error {
	if err := config.DB.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.MapData{},
		&models.Artifact{},
		&models.WorkerNode{},
	); err != nil {
		return err
	}

	if err := config.DB.AutoMigrate(
		&models.SimulationProject{},
		&models.SimulationConfig{},
		&models.SimulationResult{},
		&models.Simulation{},
	); err != nil {
		return err
	}

	return config.DB.AutoMigrate(
		&models.SimulationTask{},
		&models.TaskLog{},
	)
}
