package models

import "time"

// User 管理员账号表（仅管理员）
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	WorkNo    string    `gorm:"size:50;uniqueIndex;not null" json:"workNo"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Password  string    `gorm:"size:200" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
