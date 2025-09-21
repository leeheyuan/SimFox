package models

import "time"

type SimulationTask struct {
	ID        uint   `gorm:"primaryKey"`
	ProjectID uint   `gorm:"not null;index"`
	ConfigID  uint   `gorm:"not null;index"`
	Status    string `gorm:"size:20"` // pending, running, finished, error
	StartedAt *time.Time
	EndedAt   *time.Time
	Project   SimulationProject `gorm:"foreignKey:ProjectID"`
	Config    SimulationConfig  `gorm:"foreignKey:ConfigID"`
}
