package models

import "time"

// Question 考题
type Question struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CategoryID  uint      `gorm:"index;not null" json:"categoryId"`
	Type        string    `gorm:"size:20;not null;default:single" json:"type"`
	Title       string    `gorm:"type:text;not null" json:"title"`
	Options     string    `gorm:"type:text" json:"options"`
	Answer      string    `gorm:"size:200" json:"answer"`
	Explanation string    `gorm:"type:text" json:"explanation"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
