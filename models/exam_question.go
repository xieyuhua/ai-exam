package models

// ExamQuestion 考试-考题关联
type ExamQuestion struct {
	ExamID     uint    `gorm:"primaryKey" json:"examId"`
	QuestionID uint    `gorm:"primaryKey" json:"questionId"`
	Score      float64 `gorm:"default:0" json:"score"`
}

func (ExamQuestion) TableName() string {
	return "exam_questions"
}
