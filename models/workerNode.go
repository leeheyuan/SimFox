package models

import "time"

type WorkerNode struct {
	ID             uint       `gorm:"primaryKey"`
	TenantID       uint       `gorm:"not null;index"`
	Name           string     `gorm:"size:120;not null;index"`
	SecretHash     string     `gorm:"size:255"`
	Status         string     `gorm:"size:30;not null;index"` // online, busy, draining, offline
	Address        string     `gorm:"size:255"`
	QueueName      string     `gorm:"size:80;index"`
	LabelsJSON     string     `gorm:"type:text"`
	MaxConcurrency int        `gorm:"not null;default:1"`
	RunningTasks   int        `gorm:"not null;default:0"`
	LastHeartbeat  *time.Time `gorm:"index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Tenant Tenant `gorm:"foreignKey:TenantID"`
}
