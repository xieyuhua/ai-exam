package response

import (
	"exam-system/dto"
	"exam-system/models"
)

// SurveyWithDetail 问卷详情
type SurveyWithDetail struct {
	models.Survey
	Questions []SurveyQuestionDetail `json:"questions"`
}

// SurveyQuestionDetail 问卷题目详情
type SurveyQuestionDetail struct {
	models.SurveyQuestion
	Options []dto.RichOption `json:"options"`
}

// StudentSurveyItem 员工端问卷列表项
type StudentSurveyItem struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	AllowRepeat   bool   `json:"allowRepeat"`
	HasCompleted  bool   `json:"hasCompleted"`
	QuestionCount int    `json:"questionCount"`
}

// SurveyStatistics 问卷统计数据
type SurveyStatistics struct {
	SurveyID       uint                 `json:"surveyId"`
	TotalCompleted int64                `json:"totalCompleted"`
	TotalResponses int64                `json:"totalResponses"`
	Questions      []SurveyQuestionStat `json:"questions"`
}

// SurveyQuestionStat 单题统计
type SurveyQuestionStat struct {
	ID            uint                    `json:"id"`
	Title         string                  `json:"title"`
	Type          string                  `json:"type"`
	Required      bool                    `json:"required"`
	ResponseCount int                     `json:"responseCount"`
	Options       []SurveyOptionStatItem  `json:"options"`
	TextResponses []string                `json:"textResponses,omitempty"`
}

// SurveyOptionStatItem 问卷选项统计项
type SurveyOptionStatItem struct {
	Label   string  `json:"label"`
	Type    string  `json:"type"`
	Content string  `json:"content"`
	URL     string  `json:"url,omitempty"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}
