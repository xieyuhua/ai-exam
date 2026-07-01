package request

import "exam-system/dto"

// CreateSurveyReq 创建问卷请求
type CreateSurveyReq struct {
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description"`
	StartTime   string                 `json:"startTime" binding:"required"`
	EndTime     string                 `json:"endTime" binding:"required"`
	AllowRepeat *bool                  `json:"allowRepeat"`
	Questions   []CreateSurveyQuestion `json:"questions" binding:"required"`
}

// CreateSurveyQuestion 创建问卷题目
type CreateSurveyQuestion struct {
	Title     string         `json:"title" binding:"required"`
	Type      string         `json:"type" binding:"required"`
	Required  *bool          `json:"required"`
	SortOrder int            `json:"sortOrder"`
	Options   []dto.RichOption `json:"options"`
}

// SubmitSurveyReq 提交问卷请求
type SubmitSurveyReq struct {
	SurveyID uint                `json:"surveyId" binding:"required"`
	Answers  []SurveyAnswerItem  `json:"answers" binding:"required"`
}

// SurveyAnswerItem 问卷答案项
type SurveyAnswerItem struct {
	SurveyQuestionID uint     `json:"surveyQuestionId" binding:"required"`
	Answer           []string `json:"answer"`
}
