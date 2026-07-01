package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	dto "exam-system/dto"
	req "exam-system/dto/request"
	resp "exam-system/dto/response"
	"exam-system/models"
	"exam-system/service"
	"exam-system/util"
)

// ==================== ScoreHandler 成绩管理 + 员工端 ====================

type ScoreHandler struct {
	baseHandler
}

func NewScoreHandler(sr *service.ServiceRegistry, db *gorm.DB) *ScoreHandler {
	return &ScoreHandler{baseHandler: baseHandler{svc: sr, db: db}}
}

// ==================== 成绩管理（管理员） ====================

// GetScores 获取成绩列表
// @Summary      获取成绩列表
// @Description  分页查询所有员工的考试成绩，支持按关键词、分类筛选
// @Tags         成绩管理
// @Accept       json
// @Produce      json
// @Param        keyword     query  string  false  "搜索关键词（员工姓名/考试名称）"
// @Param        categoryId  query  string  false  "分类 ID"
// @Param        page        query  int     false  "页码 (默认 1)"
// @Param        pageSize    query  int     false  "每页条数 (默认 10)"
// @Success      200         {object}  resp.APIResponse{data=resp.PagedData}  "成绩列表"
// @Failure      500         {object}  resp.APIResponse  "查询失败"
// @Security     BearerToken
// @Router       /api/admin/scores [get]
func (h *ScoreHandler) GetScores(c *gin.Context) {
	page, pageSize := util.ParsePagination(c.Query("page"), c.Query("pageSize"))
	data, err := h.svc.Score.List(c.Query("keyword"), c.Query("categoryId"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// ExportScores 导出成绩
// @Summary      导出成绩
// @Description  按分类导出成绩为 CSV 或 Excel 文件
// @Tags         成绩管理
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,text/csv
// @Param        categoryId  query  string  false  "分类 ID（为空则导出所有）"
// @Param        format      query  string  false  "导出格式：csv 或 xlsx（默认 xlsx）"
// @Success      200         {file}  binary  "成绩文件下载"
// @Failure      500         {object}  resp.APIResponse  "导出失败"
// @Security     BearerToken
// @Router       /api/admin/scores/export [get]
func (h *ScoreHandler) ExportScores(c *gin.Context) {
	categoryID := c.Query("categoryId")
	format := strings.ToLower(c.DefaultQuery("format", "xlsx"))

	rows, err := h.svc.Score.BuildExportRows(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "导出失败"})
		return
	}

	headers := []string{"学号", "考试分类", "考试名称", "提交日期", "得分", "正确数", "总题数", "正确率(%)"}

	if format == "csv" {
		h.exportCSV(c, headers, rows)
	} else {
		h.exportExcel(c, headers, rows)
	}
}

// ExportExamScores 导出指定考试的成绩
func (h *ScoreHandler) ExportExamScores(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的考试ID"})
		return
	}
	format := strings.ToLower(c.DefaultQuery("format", "xlsx"))
	db := h.getDB()

	var exam models.Exam
	if err := db.First(&exam, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "考试不存在"})
		return
	}

	rows, err := h.svc.Score.BuildExportRowsByExam(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "导出失败"})
		return
	}

	headers := []string{"学号", "考试分类", "考试名称", "提交日期", "得分", "正确数", "总题数", "正确率(%)"}
	safeTitle := strings.ReplaceAll(exam.Title, "/", "_")

	if format == "csv" {
		filename := fmt.Sprintf("成绩_%s_%s.csv", safeTitle, time.Now().Format("20060102_150405"))
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
		writer := csv.NewWriter(c.Writer)
		writer.Write(headers)
		for _, row := range rows {
			rate := fmt.Sprintf("%.1f", float64(row.Correct)/float64(row.Total)*100)
			writer.Write([]string{
				row.StudentName, row.CategoryName, row.ExamTitle,
				row.Date, fmt.Sprintf("%.1f", row.Score),
				strconv.Itoa(row.Correct), strconv.Itoa(row.Total), rate,
			})
		}
		writer.Flush()
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "考试成绩"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"6366F1"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	for i, hdr := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hdr)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheet, 1, 25)

	scoreStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "F5576C"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	for i, row := range rows {
		rowNum := i + 2
		rate := fmt.Sprintf("%.1f%%", float64(row.Correct)/float64(row.Total)*100)
		values := []interface{}{row.StudentName, row.CategoryName, row.ExamTitle, row.Date, row.Score, row.Correct, row.Total, rate}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowNum)
			f.SetCellValue(sheet, cell, v)
		}
		scoreCell, _ := excelize.CoordinatesToCellName(5, rowNum)
		f.SetCellStyle(sheet, scoreCell, scoreCell, scoreStyle)
	}
	f.SetColWidth(sheet, "A", "A", 14)
	f.SetColWidth(sheet, "B", "B", 14)
	f.SetColWidth(sheet, "C", "C", 24)
	f.SetColWidth(sheet, "D", "D", 18)
	f.SetColWidth(sheet, "E", "E", 12)
	f.SetColWidth(sheet, "F", "F", 10)
	f.SetColWidth(sheet, "G", "G", 10)
	f.SetColWidth(sheet, "H", "H", 12)

	filename := fmt.Sprintf("成绩_%s_%s.xlsx", safeTitle, time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

func (h *ScoreHandler) exportCSV(c *gin.Context, headers []string, rows []service.ScoreExportRow) {
	filename := fmt.Sprintf("成绩导出_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(c.Writer)
	writer.Write(headers)

	for _, row := range rows {
		rate := fmt.Sprintf("%.1f", float64(row.Correct)/float64(row.Total)*100)
		writer.Write([]string{
			row.StudentName,
			row.CategoryName,
			row.ExamTitle,
			row.Date,
			fmt.Sprintf("%.1f", row.Score),
			strconv.Itoa(row.Correct),
			strconv.Itoa(row.Total),
			rate,
		})
	}
	writer.Flush()
}

func (h *ScoreHandler) exportExcel(c *gin.Context, headers []string, rows []service.ScoreExportRow) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "成绩表"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F5576C"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	for i, hdr := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hdr)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheet, 1, 25)

	scoreStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "F5576C"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	for i, row := range rows {
		rowNum := i + 2
		rate := fmt.Sprintf("%.1f%%", float64(row.Correct)/float64(row.Total)*100)
		values := []interface{}{row.StudentName, row.CategoryName, row.ExamTitle, row.Date, row.Score, row.Correct, row.Total, rate}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowNum)
			f.SetCellValue(sheet, cell, v)
		}
		scoreCell, _ := excelize.CoordinatesToCellName(5, rowNum)
		f.SetCellStyle(sheet, scoreCell, scoreCell, scoreStyle)
	}

	f.SetColWidth(sheet, "A", "A", 14)
	f.SetColWidth(sheet, "B", "B", 14)
	f.SetColWidth(sheet, "C", "C", 24)
	f.SetColWidth(sheet, "D", "D", 18)
	f.SetColWidth(sheet, "E", "E", 12)
	f.SetColWidth(sheet, "F", "F", 10)
	f.SetColWidth(sheet, "G", "G", 10)
	f.SetColWidth(sheet, "H", "H", 12)

	filename := fmt.Sprintf("成绩导出_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

// ==================== 员工端 ====================

// GetStudentExams 获取可用考试列表（员工端）
// @Summary      获取可用考试列表
// @Description  员工端查看当前可参加的考试（状态为 active/upcoming），含是否已参加的标记
// @Tags         员工端
// @Accept       json
// @Produce      json
// @Success      200  {object}  resp.APIResponse{data=[]resp.StudentExamItem}  "考试列表"
// @Security     BearerToken
// @Router       /api/student/exams [get]
func (h *ScoreHandler) GetStudentExams(c *gin.Context) {
	db := h.getDB()

	var studentName string
	if ident := GetCurrentIdentity(c); ident != nil {
		studentName = ident.Name
	}

	var exams []models.Exam
	db.Where("status IN ?", []string{"active", "upcoming"}).Order("start_time asc").Find(&exams)

	now := time.Now()
	loc := now.Location()
	result := make([]resp.StudentExamItem, len(exams))
	for i, e := range exams {
		// DB (SQLite) 不存时区，读出被当 UTC；这里按本地时区重新解释后比较
		startTime := time.Date(e.StartTime.Year(), e.StartTime.Month(), e.StartTime.Day(),
			e.StartTime.Hour(), e.StartTime.Minute(), e.StartTime.Second(), 0, loc)
		endTime := time.Date(e.EndTime.Year(), e.EndTime.Month(), e.EndTime.Day(),
			e.EndTime.Hour(), e.EndTime.Minute(), e.EndTime.Second(), 0, loc)

		// 实时计算考试状态（而非依赖数据库静态字段）
		realStatus := "upcoming"
		if now.After(startTime) && now.Before(endTime) {
			realStatus = "active"
		} else if now.After(endTime) {
			realStatus = "ended"
		}

		// 实时查询考题数量和分值总和
		var eqs []models.ExamQuestion
		db.Where("exam_id = ?", e.ID).Find(&eqs)
		qCount := len(eqs)
		actualScore := 0.0
		for _, eq := range eqs {
			actualScore += eq.Score
		}
		totalScore := int(actualScore)
		if totalScore == 0 {
			totalScore = e.TotalScore
		}

		allowRepeat := false
		if e.AllowRepeat != nil {
			allowRepeat = *e.AllowRepeat
		}
		canView := true
		if e.CanViewAnswer != nil {
			canView = *e.CanViewAnswer
		}

		hasAttempted := false
		if studentName != "" {
			var count int64
			db.Model(&models.Score{}).Where("exam_id = ? AND student_name = ?", e.ID, studentName).Count(&count)
			hasAttempted = count > 0
		}

		result[i] = resp.StudentExamItem{
			ID:            e.ID,
			Title:         e.Title,
			QuestionCount: qCount,
			TotalScore:    totalScore,
			Duration:      e.Duration,
			Status:        realStatus,
			StartTime:     startTime.Format("2006-01-02 15:04"),
			EndTime:       endTime.Format("2006-01-02 15:04"),
			AllowRepeat:   allowRepeat,
			CanViewAnswer: canView,
			HasAttempted:  hasAttempted,
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// GetExamQuestions 获取考试题目（员工端）
// @Summary      获取考试题目
// @Description  员工端获取考试题目（不含答案）。如果已提交且不允许重复，则返回历史成绩
// @Tags         员工端
// @Accept       json
// @Produce      json
// @Param        examId       query  int     true   "考试 ID"
// @Param        studentName  query  string  false  "员工姓名"
// @Success      200          {object}  resp.APIResponse{data=resp.ExamQuestionsData}  "考题列表或历史成绩"
// @Failure      400          {object}  resp.APIResponse  "缺少 examId"
// @Failure      404          {object}  resp.APIResponse  "考试不存在"
// @Security     BearerToken
// @Router       /api/exam/questions [get]
func (h *ScoreHandler) GetExamQuestions(c *gin.Context) {
	examIDStr := c.Query("examId")
	studentName := c.Query("studentName")
	if examIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少 examId 参数"})
		return
	}
	examID, err := strconv.ParseUint(examIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的 examId"})
		return
	}

	db := h.getDB()
	uid := uint(examID)

	var exam models.Exam
	if db.First(&exam, uid).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "考试不存在"})
		return
	}

	// 实时校验考试时间状态（DB 无时区，需按本地时区重新解释）
	now := time.Now()
	loc := now.Location()
	startTime := time.Date(exam.StartTime.Year(), exam.StartTime.Month(), exam.StartTime.Day(),
		exam.StartTime.Hour(), exam.StartTime.Minute(), exam.StartTime.Second(), 0, loc)
	endTime := time.Date(exam.EndTime.Year(), exam.EndTime.Month(), exam.EndTime.Day(),
		exam.EndTime.Hour(), exam.EndTime.Minute(), exam.EndTime.Second(), 0, loc)
	if now.Before(startTime) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "考试尚未开始"})
		return
	}
	if now.After(endTime) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "考试已结束"})
		return
	}

	canView := true
	if exam.CanViewAnswer != nil {
		canView = *exam.CanViewAnswer
	}
	repeat := false
	if exam.AllowRepeat != nil {
		repeat = *exam.AllowRepeat
	}

	// 检查是否已提交过
	if !repeat && studentName != "" {
		score, err := h.svc.Score.GetByExamAndStudent(uid, studentName)
		if err == nil {
			var results []resp.ResultItem
			json.Unmarshal([]byte(score.Answers), &results)
			unanswered := score.Total - score.Correct
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{
				"isCompleted":    true,
				"canViewAnswer":  canView,
				"allowRepeat":    repeat,
				"title":          exam.Title,
				"totalScore":     score.Score,
				"correctCount":   score.Correct,
				"wrongCount":     unanswered,
				"totalQuestions": score.Total,
				"results":        results,
				"list":           []interface{}{},
			}})
			return
		}
	}

	var eqs []models.ExamQuestion
	db.Where("exam_id = ?", uid).Find(&eqs)

	if len(eqs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"examId": exam.ID, "title": exam.Title, "totalQuestions": 0, "list": []interface{}{}}})
		return
	}

	questionIDs := make([]uint, len(eqs))
	for i, eq := range eqs {
		questionIDs[i] = eq.QuestionID
	}

	var questions []models.Question
	db.Where("id IN ?", questionIDs).Find(&questions)
	qMap := make(map[uint]models.Question)
	for _, q := range questions {
		qMap[q.ID] = q
	}

	list := make([]resp.ExamQuestionItem, 0)
	for _, eq := range eqs {
		q, ok := qMap[eq.QuestionID]
		if !ok {
			continue
		}
		list = append(list, resp.ExamQuestionItem{
			ID:      q.ID,
			ExamID:  exam.ID,
			Type:    q.Type,
			Title:   q.Title,
			Options: h.parseOptionsSorted(q.Options),
			Score:   eq.Score,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{
		"isCompleted":    false,
		"canViewAnswer":  canView,
		"allowRepeat":    repeat,
		"examId":         exam.ID,
		"title":          exam.Title,
		"totalQuestions": len(list),
		"list":           list,
	}})
}

// SubmitExam 提交考试
// @Summary      提交考试答案
// @Description  员工提交考试答案，系统自动判分并返回结果
// @Tags         员工端
// @Accept       json
// @Produce      json
// @Param        body  body  req.SubmitExamReq  true  "答案数据"
// @Success      200   {object}  resp.APIResponse{data=resp.ExamResultRes}  "判分结果"
// @Failure      400   {object}  resp.APIResponse  "参数错误或不允许重复提交"
// @Security     BearerToken
// @Router       /api/exam/submit [post]
func (h *ScoreHandler) SubmitExam(c *gin.Context) {
	var r req.SubmitExamReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	db := h.getDB()

	var eqs []models.ExamQuestion
	db.Where("exam_id = ?", r.ExamID).Find(&eqs)

	questionIDs := make([]uint, len(eqs))
	eqMap := make(map[uint]bool)
	scoreMap := make(map[uint]float64)
	for i, eq := range eqs {
		questionIDs[i] = eq.QuestionID
		eqMap[eq.QuestionID] = true
		scoreMap[eq.QuestionID] = eq.Score
	}

	var questions []models.Question
	db.Where("id IN ?", questionIDs).Find(&questions)
	qMap := make(map[uint]models.Question)
	for _, q := range questions {
		qMap[q.ID] = q
	}

	userAnswerMap := make(map[uint][]string)
	for _, a := range r.Answers {
		if eqMap[a.QuestionID] {
			userAnswerMap[a.QuestionID] = a.SelectedOptions
		}
	}

	correctCount := 0
	wrongCount := 0
	unansweredCount := 0
	totalQuestionCount := len(questionIDs)
	var earnedScore float64

	results := make([]resp.ResultItem, 0)
	for _, qid := range questionIDs {
		q, ok := qMap[qid]
		if !ok {
			continue
		}

		sc := scoreMap[qid]
		if sc == 0 {
			sc = 10
		}

		userAns := userAnswerMap[qid]
		if userAns == nil || len(userAns) == 0 {
			unansweredCount++
			var correctAns []string
			if err := json.Unmarshal([]byte(q.Answer), &correctAns); err != nil {
				correctAns = []string{}
			}
			results = append(results, resp.ResultItem{
				QuestionID: qid, IsCorrect: false, IsUnanswered: true,
				UserAnswer: []string{}, CorrectAnswer: correctAns,
			})
			continue
		}

		var correctAns []string
		if err := json.Unmarshal([]byte(q.Answer), &correctAns); err != nil {
			correctAns = []string{}
		}

		isCorrect := false
		if q.Type == "fill" || q.Type == "essay" {
			userAnsText := strings.TrimSpace(userAns[0])
			correctAnsText := ""
			if len(correctAns) > 0 {
				correctAnsText = strings.TrimSpace(correctAns[0])
			}
			isCorrect = userAnsText != "" && strings.EqualFold(userAnsText, correctAnsText)
		} else {
			isCorrect = util.StringSliceEqual(util.SortSlice(userAns), util.SortSlice(correctAns))
		}

		if isCorrect {
			correctCount++
			earnedScore += sc
		} else {
			wrongCount++
		}

		results = append(results, resp.ResultItem{
			QuestionID: qid, IsCorrect: isCorrect, IsUnanswered: false,
			UserAnswer: userAns, CorrectAnswer: correctAns,
		})
	}

	// 按题目实际分值累加计算总得分
	totalScore := math.Round(earnedScore*10) / 10

	// 检查允许重复
	var exam models.Exam
	db.First(&exam, r.ExamID)
	allowRepeat := false
	if exam.AllowRepeat != nil {
		allowRepeat = *exam.AllowRepeat
	}
	if !allowRepeat {
		count := h.svc.Score.CountByExamAndStudent(r.ExamID, r.StudentName)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该考试不允许重复参加"})
			return
		}
	}

	canView := true
	if exam.CanViewAnswer != nil {
		canView = *exam.CanViewAnswer
	}

	answersJSON, _ := json.Marshal(results)
	score := models.Score{
		StudentName: r.StudentName,
		ExamID:      r.ExamID,
		Score:       totalScore,
		Correct:     correctCount,
		Total:       totalQuestionCount,
		Answers:     string(answersJSON),
	}
	h.svc.Score.Save(&score)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": resp.ExamResultRes{
		TotalScore:      totalScore,
		CorrectCount:    correctCount,
		WrongCount:      wrongCount,
		UnansweredCount: unansweredCount,
		TotalQuestions:  totalQuestionCount,
		CanViewAnswer:   canView,
		Results:         results,
	}})
}

// GetStudentRecords 获取员工成绩记录
// @Summary      获取员工历史成绩
// @Description  查询指定员工的所有考试成绩记录
// @Tags         员工端
// @Accept       json
// @Produce      json
// @Param        studentName  query  string  true  "员工姓名"
// @Success      200          {object}  resp.APIResponse{data=[]resp.StudentRecordItem}  "成绩列表"
// @Failure      500          {object}  resp.APIResponse  "查询失败"
// @Security     BearerToken
// @Router       /api/student/records [get]
func (h *ScoreHandler) GetStudentRecords(c *gin.Context) {
	studentName := c.Query("studentName")
	scores, err := h.svc.Score.ListByStudent(studentName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	db := h.getDB()
	result := make([]gin.H, len(scores))
	for i, s := range scores {
		var exam models.Exam
		db.First(&exam, s.ExamID)

		canView := true
		if exam.CanViewAnswer != nil {
			canView = *exam.CanViewAnswer
		}

		result[i] = gin.H{
			"id":            s.ID,
			"examTitle":     exam.Title,
			"date":          s.CreatedAt.Local().Format("2006-01-02 15:04"),
			"score":         s.Score,
			"correctCount":  s.Correct,
			"totalCount":    s.Total,
			"canViewAnswer": canView,
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// ==================== 私有工具 ====================

func (h *ScoreHandler) parseOptionsSorted(optionsJSON string) []dto.OptionPair {
	return service.ParseOptions(optionsJSON)
}
