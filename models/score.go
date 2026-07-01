package models

import "time"

// Score 成绩
type Score struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	StudentName string    `gorm:"size:100;not null" json:"studentName"`
	ExamID      uint      `gorm:"index;not null" json:"examId"`
	Score       float64   `json:"score"`
	Correct     int       `json:"correct"`
	Total       int       `json:"total"`
	Answers     string    `gorm:"type:text" json:"answers"`
	CreatedAt   time.Time `json:"createdAt"`
}
