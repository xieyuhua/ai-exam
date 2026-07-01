package response

import "exam-system/models"

// APIResponse 通用 API 响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// PagedData 分页响应结构体
type PagedData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// BatchImportResult 批量导入结果
type BatchImportResult struct {
	Added   int `json:"added"`
	Skipped int `json:"skipped"`
}

// CategoryWithCount 带计数的分类（用于列表展示）
type CategoryWithCount struct {
	models.Category
	ExamCount     int `json:"examCount"`
	QuestionCount int `json:"questionCount"`
}
