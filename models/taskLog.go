package models

import "time"

type TaskLog struct {
	ID        uint      `gorm:"primaryKey"`
	TaskID    uint      `gorm:"not null;index"`
	WorkerID  *uint     `gorm:"index"`
	Level     string    `gorm:"size:20;not null;default:info"`
	Message   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"index"`
}
