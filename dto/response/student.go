package response

// StudentRecordItem 员工成绩记录项
type StudentRecordItem struct {
	ID            uint    `json:"id"`
	ExamTitle     string  `json:"examTitle"`
	Date          string  `json:"date"`
	Score         float64 `json:"score"`
	CorrectCount  int     `json:"correctCount"`
	TotalCount    int     `json:"totalCount"`
	CanViewAnswer bool    `json:"canViewAnswer"`
}
