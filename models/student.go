package models

import (
	"time"

	"gorm.io/gorm"
)

// Student 员工表（独立于管理员）
type Student struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	WorkNo      string    `gorm:"size:50;uniqueIndex;not null" json:"workNo"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Password    string    `gorm:"size:200" json:"-"`
	WxUserID    string    `gorm:"size:100;index" json:"wxUserId"`
	WxOpenID    string    `gorm:"size:200" json:"wxOpenId"`
	Avatar      string    `gorm:"size:500" json:"avatar"`
	Source      string    `gorm:"size:20;default:import" json:"source"`
	HasPassword bool      `gorm:"-" json:"hasPassword"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// AfterFind GORM 查询回调：自动计算 HasPassword
func (s *Student) AfterFind(tx *gorm.DB) error {
	s.HasPassword = s.Password != ""
	return nil
}
