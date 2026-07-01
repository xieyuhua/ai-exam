package response

import "exam-system/models"

// ScoreWithDetail 带考试信息的成绩记录
type ScoreWithDetail struct {
	models.Score
	ExamTitle     string `json:"examTitle"`
	CategoryName  string `json:"categoryName"`
	CanViewAnswer bool   `json:"canViewAnswer"`
}

// ScoreExportRowData 成绩导出行
type ScoreExportRowData struct {
	StudentName  string  `json:"studentName"`
	CategoryName string  `json:"categoryName"`
	ExamTitle    string  `json:"examTitle"`
	Date         string  `json:"date"`
	Score        float64 `json:"score"`
	Correct      int     `json:"correct"`
	Total        int     `json:"total"`
	Rate         string  `json:"rate"`
}

// StatisticsData 管理端总览统计
type StatisticsData struct {
	TotalExams     int64 `json:"totalExams"`
	TotalQuestions int64 `json:"totalQuestions"`
	TotalStudents  int64 `json:"totalStudents"`
	TotalScores    int64 `json:"totalScores"`
	TotalVotes     int64 `json:"totalVotes"`
	TotalSurveys   int64 `json:"totalSurveys"`
	ActiveExams    int64 `json:"activeExams"`
	ActiveVotes    int64 `json:"activeVotes"`
	ActiveSurveys  int64 `json:"activeSurveys"`
}
