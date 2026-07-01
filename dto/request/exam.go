package request

// CreateExamReq 创建考试请求
type CreateExamReq struct {
	CategoryID    uint   `json:"categoryId" binding:"required"`
	Title         string `json:"title" binding:"required"`
	Duration      int    `json:"duration" binding:"required"`
	TotalScore    int    `json:"totalScore" binding:"required"`
	StartTime     string `json:"startTime" binding:"required"`
	EndTime       string `json:"endTime" binding:"required"`
	CanViewAnswer *bool  `json:"canViewAnswer"`
	AllowRepeat   *bool  `json:"allowRepeat"`
	QuestionIDs   []uint `json:"questionIds"`
}

// UpdateExamReq 更新考试请求
type UpdateExamReq struct {
	CategoryID    uint   `json:"categoryId"`
	Title         string `json:"title"`
	Duration      int    `json:"duration"`
	TotalScore    int    `json:"totalScore"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	CanViewAnswer *bool  `json:"canViewAnswer"`
	AllowRepeat   *bool  `json:"allowRepeat"`
	QuestionIDs   []uint `json:"questionIds"`
}

// SubmitExamReq 提交考试请求
type SubmitExamReq struct {
	ExamID      uint         `json:"examId" binding:"required"`
	StudentName string       `json:"studentName" binding:"required"`
	Answers     []AnswerItem `json:"answers"`
	SubmitTime  string       `json:"submitTime"`
}

// AnswerItem 答题项
type AnswerItem struct {
	QuestionID      uint     `json:"questionId"`
	SelectedOptions []string `json:"selectedOptions"`
}

// ImportExamQuestionsReq 导入考题到考试请求
type ImportExamQuestionsReq struct {
	QuestionIDs  []uint           `json:"questionIds" binding:"required"`
	DefaultScore float64          `json:"defaultScore"`
	Scores       map[uint]float64 `json:"scores"`
}
