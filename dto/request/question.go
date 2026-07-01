package request

import "exam-system/dto"

// CreateQuestionReq 创建题目请求
type CreateQuestionReq struct {
	CategoryID  uint           `json:"categoryId" binding:"required"`
	Type        string         `json:"type" binding:"required"`
	Title       string         `json:"title" binding:"required"`
	Options     []dto.RichOption `json:"options"`
	Answer      []string       `json:"answer" binding:"required"`
	Explanation string         `json:"explanation"`
}

// UpdateQuestionReq 更新题目请求
type UpdateQuestionReq struct {
	CategoryID  uint           `json:"categoryId"`
	Type        string         `json:"type"`
	Title       string         `json:"title"`
	Options     []dto.RichOption `json:"options"`
	Answer      []string       `json:"answer"`
	Explanation string         `json:"explanation"`
}
