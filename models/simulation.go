package models

import "time"

type Simulation struct {
	ID         uint   `gorm:"primaryKey"`
	ProjectID  uint   `gorm:"not null;index"`
	UserID     uint   `gorm:"not null;index"`
	Name       string `gorm:"size:100"`
	ConfigJSON string `gorm:"type:text"` // 存储 JSON 字符串
	CreatedAt  time.Time
	Project    SimulationProject `gorm:"foreignKey:ProjectID"`
	User       User              `gorm:"foreignKey:UserID"`
}
