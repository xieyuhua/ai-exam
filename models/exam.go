package models

import "time"

// Exam 考试
type Exam struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CategoryID    uint      `gorm:"index;not null" json:"categoryId"`
	Title         string    `gorm:"size:200;not null" json:"title"`
	Duration      int       `gorm:"default:30" json:"duration"`
	TotalScore    int       `gorm:"default:100" json:"totalScore"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	CanViewAnswer *bool     `gorm:"default:true" json:"canViewAnswer"`
	AllowRepeat   *bool     `gorm:"default:false" json:"allowRepeat"`
	Status        string    `gorm:"size:20;default:upcoming" json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
