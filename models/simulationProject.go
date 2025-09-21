package models

import "time"

type SimulationProject struct {
	ID          uint    `gorm:"primaryKey"`
	TenantID    uint    `gorm:"not null;index"`
	Name        string  `gorm:"size:100;not null"`
	Description string  `gorm:"size:500"`
	MapDataID   uint    // 外键字段
	MapData     MapData `gorm:"foreignKey:MapDataID"`
	Status      string  `gorm:"size:20"` // pending, running, finished, error
	CreatedAt   time.Time
	UpdateAt    time.Time
	Tenant      Tenant `gorm:"foreignKey:TenantID"`
}
