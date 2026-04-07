package models

import "time"

type SimulationTask struct {
	ID              uint    `gorm:"primaryKey"`
	ProjectID       uint    `gorm:"not null;index"`
	ConfigID        uint    `gorm:"not null;index"`
	Status          string  `gorm:"size:20;index"` // pending, running, finished, error
	DurationSeconds int32   `gorm:"not null;default:60"`
	Speed           float64 `gorm:"not null;default:1"`
	MonitorPort     int
	TraCIPort       int
	LastError       string `gorm:"size:500"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	StartedAt       *time.Time
	EndedAt         *time.Time
	Project         SimulationProject `gorm:"foreignKey:ProjectID"`
	Config          SimulationConfig  `gorm:"foreignKey:ConfigID"`
}
