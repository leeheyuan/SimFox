package models

import "time"

type SimulationRecentlyOpened struct {
	ID          uint    `gorm:"primaryKey"`
	TenantID    uint    `gorm:"not null;index"` 
	ProjectID   uint    `gorm:"not null;index"`
	Project     SimulationProject `gorm:"foreignKey:ProjectID"`
	OpenedAt    time.Time
	UpdateAt    time.Time
	Tenant      Tenant `gorm:"foreignKey:TenantID"`
}
