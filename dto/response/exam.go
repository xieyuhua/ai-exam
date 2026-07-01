package response

import (
	"exam-system/dto"
	"exam-system/models"
)

// ExamDetailData 考试详情（含考题列表）
type ExamDetailData struct {
	Exam          models.Exam           `json:"exam"`
	Questions     []ExamQuestionDetail  `json:"questions"`
	TotalScore    float64               `json:"totalScore"`
	QuestionCount int                   `json:"questionCount"`
}

// ExamQuestionDetail 考试内考题详情（含分值）
type ExamQuestionDetail struct {
	QuestionID  uint            `json:"questionId"`
	CategoryID  uint            `json:"categoryId"`
	Type        string          `json:"type"`
	Title       string          `json:"title"`
	Options     []dto.OptionPair `json:"options"`
	Answer      []string        `json:"answer"`
	Explanation string          `json:"explanation"`
	Score       float64         `json:"score"`
}

// ExamWithDetail 考试完整信息（含分类名、题目数）
type ExamWithDetail struct {
	models.Exam
	CategoryName string `json:"categoryName"`
	QuestionCount int   `json:"questionCount"`
}
