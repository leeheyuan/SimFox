package models

import "time"

type Artifact struct {
	ID          uint   `gorm:"primaryKey"`
	TenantID    uint   `gorm:"not null;index"`
	ProjectID   *uint  `gorm:"index"`
	TaskID      *uint  `gorm:"index"`
	Kind        string `gorm:"size:40;not null;index"` // input, output, log, report
	Name        string `gorm:"size:160;not null"`
	FileURL     string `gorm:"size:500;not null"`
	ContentType string `gorm:"size:120"`
	SizeBytes   int64
	Checksum    string `gorm:"size:128"`
	CreatedAt   time.Time
}
