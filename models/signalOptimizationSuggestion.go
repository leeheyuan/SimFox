package models

import "time"

type SignalOptimizationSuggestion struct {
	ID                 uint   `gorm:"primaryKey"`
	TaskID             uint   `gorm:"not null;uniqueIndex"`
	ProjectID          uint   `gorm:"not null;index"`
	TenantID           uint   `gorm:"not null;index"`
	RerunTaskID        *uint  `gorm:"index"`
	Status             string `gorm:"size:20;not null;default:'pending'"`
	Engine             string `gorm:"size:40;not null;default:'heuristic'"`
	Summary            string `gorm:"size:1000"`
	MetricsJSON        string `gorm:"type:text"`
	AnalysisJSON       string `gorm:"type:text"`
	ProposalJSON       string `gorm:"type:text"`
	ProposedConfigPath string `gorm:"size:500"`
	AppliedConfigPath  string `gorm:"size:500"`
	AppliedSignalFile  string `gorm:"size:500"`
	RejectedReason     string `gorm:"size:500"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ReviewedAt         *time.Time
	Task               SimulationTask    `gorm:"foreignKey:TaskID"`
	Project            SimulationProject `gorm:"foreignKey:ProjectID"`
	Tenant             Tenant            `gorm:"foreignKey:TenantID"`
}
