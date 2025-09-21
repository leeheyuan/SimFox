package models

import "time"

type SimulationResult struct {
	ID          uint   `gorm:"primaryKey"`
	TaskID      uint   `gorm:"not null;index"`
	FileURL     string `gorm:"size:255"`  // 结果文件（如 CSV、JSON）
	Metrics     string `gorm:"type:text"` // 指标 JSON
	GeneratedAt time.Time

	Task SimulationTask `gorm:"foreignKey:TaskID"`
}
