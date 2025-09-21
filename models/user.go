package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"not null;index"`
	Username  string `gorm:"size:100;not null;uniqueIndex"`
	Password  string `gorm:"size:255;not null"` // 存储哈希后的密码
	Role      string `gorm:"size:50"`           // 如 admin/user
	CreatedAt time.Time

	Tenant Tenant `gorm:"foreignKey:TenantID"`
}

func (u *User) SetPassword(pw string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (u *User) CheckPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pw)) == nil
}
