package models

import "time"

// Survey 调查问卷表
type Survey struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Title          string    `gorm:"size:200;not null" json:"title"`
	Description    string    `gorm:"type:text" json:"description"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`
	AllowRepeat    *bool     `gorm:"default:false" json:"allowRepeat"`
	Status         string    `gorm:"size:20;default:upcoming" json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	QuestionCount  int       `gorm:"-" json:"questionCount"`
	TotalCompleted int64     `gorm:"-" json:"totalCompleted"`
}

// SurveyQuestion 问卷题目表
type SurveyQuestion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SurveyID  uint      `gorm:"index;not null" json:"surveyId"`
	Title     string    `gorm:"type:text;not null" json:"title"`
	Type      string    `gorm:"size:20;default:single" json:"type"`
	SortOrder int       `gorm:"default:0" json:"sortOrder"`
	Required  *bool     `gorm:"default:true" json:"required"`
	CreatedAt time.Time `json:"createdAt"`
}

// SurveyOption 问卷选项表
type SurveyOption struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	SurveyQuestionID uint   `gorm:"index;not null" json:"surveyQuestionId"`
	Label            string `gorm:"size:10;not null" json:"label"`
	Type             string `gorm:"size:20;default:text" json:"type"`
	Content          string `gorm:"type:text" json:"content"`
	URL              string `gorm:"size:500" json:"url"`
	SortOrder        int    `gorm:"default:0" json:"sortOrder"`
}

// SurveyAnswer 问卷答案表
type SurveyAnswer struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	SurveyID         uint      `gorm:"index;not null" json:"surveyId"`
	SurveyQuestionID uint      `gorm:"index;not null" json:"surveyQuestionId"`
	StudentName      string    `gorm:"size:100;not null" json:"studentName"`
	Answer           string    `gorm:"type:text" json:"answer"`
	CreatedAt        time.Time `json:"createdAt"`
}
