package models

import "time"

type Tenant struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null;uniqueIndex"`
	CreatedAt time.Time
}
