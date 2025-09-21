package models

import "time"

type DamnedData struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"size:100"`
	FileURL    string `gorm:"size:255"` // 对象存储路径
	Format     string `gorm:"size:20"`  // 如 net.xml
	TenantID   uint   `gorm:"not null;index"`
	MapDataID  uint   `gorm:"not null;index"`
	UploadedAt time.Time
	Tenant     Tenant  `gorm:"foreignKey:TenantID"`
	MapData    MapData `gorm:"foreignKey:MapDataID"`
}
