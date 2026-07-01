package response

import "exam-system/dto"

// ExamQuestionsData 考试题目列表响应
type ExamQuestionsData struct {
	IsCompleted    bool                `json:"isCompleted"`
	CanViewAnswer  bool                `json:"canViewAnswer"`
	AllowRepeat    bool                `json:"allowRepeat"`
	ExamID         uint                `json:"examId"`
	Title          string              `json:"title"`
	TotalQuestions int                 `json:"totalQuestions"`
	List           []ExamQuestionItem  `json:"list"`
	TotalScore     float64             `json:"totalScore,omitempty"`
	CorrectCount   int                 `json:"correctCount,omitempty"`
	WrongCount     int                 `json:"wrongCount,omitempty"`
	Results        []ResultItem        `json:"results,omitempty"`
}

// ExamQuestionItem 考题详情（含完整选项与分值）
type ExamQuestionItem struct {
	ID          uint            `json:"id"`
	ExamID      uint            `json:"examId"`
	Type        string          `json:"type"`
	Title       string          `json:"title"`
	Options     []dto.OptionPair `json:"options"`
	Answer      []string        `json:"answer,omitempty"`
	Explanation string          `json:"explanation,omitempty"`
	Score       float64         `json:"score"`
}

// ResultItem 考试结果项
type ResultItem struct {
	QuestionID    uint     `json:"questionId"`
	IsCorrect     bool     `json:"isCorrect"`
	IsUnanswered  bool     `json:"isUnanswered"`
	UserAnswer    []string `json:"userAnswer"`
	CorrectAnswer []string `json:"correctAnswer"`
}

// StudentExamItem 员工端考试列表项
type StudentExamItem struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	QuestionCount int    `json:"questionCount"`
	TotalScore    int    `json:"totalScore"`
	Duration      int    `json:"duration"`
	Status        string `json:"status"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	AllowRepeat   bool   `json:"allowRepeat"`
	CanViewAnswer bool   `json:"canViewAnswer"`
	HasAttempted  bool   `json:"hasAttempted"`
}

// ExamResultRes 考试结果响应
type ExamResultRes struct {
	TotalScore      float64      `json:"totalScore"`
	CorrectCount    int          `json:"correctCount"`
	WrongCount      int          `json:"wrongCount"`
	UnansweredCount int          `json:"unansweredCount"`
	TotalQuestions  int          `json:"totalQuestions"`
	CanViewAnswer   bool         `json:"canViewAnswer"`
	Results         []ResultItem `json:"results,omitempty"`
}
