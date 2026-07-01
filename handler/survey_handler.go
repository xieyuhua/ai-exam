package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"exam-system/models"
	"exam-system/service"
	"exam-system/util"
	"net/http"
	"strconv"
	"strings"
	"time"

	req "exam-system/dto/request"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type SurveyHandler struct {
	baseHandler
}

func NewSurveyHandler(sr *service.ServiceRegistry, db *gorm.DB) *SurveyHandler {
	return &SurveyHandler{baseHandler: baseHandler{svc: sr, db: db}}
}

// GetSurveys 问卷列表
func (h *SurveyHandler) GetSurveys(c *gin.Context) {
	page, pageSize := util.ParsePagination(c.Query("page"), c.Query("pageSize"))
	data, err := h.svc.Survey.List(c.Query("keyword"), c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// CreateSurvey 创建问卷
func (h *SurveyHandler) CreateSurvey(c *gin.Context) {
	var reqForm req.CreateSurveyReq
	if err := c.ShouldBindJSON(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	s, err := h.svc.Survey.Create(reqForm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": s, "message": "创建成功"})
}

// GetSurveyDetail 问卷详情
func (h *SurveyHandler) GetSurveyDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	data, err := h.svc.Survey.GetDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "问卷不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// UpdateSurvey 更新问卷
func (h *SurveyHandler) UpdateSurvey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	var reqForm req.CreateSurveyReq
	if err := c.ShouldBindJSON(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := h.svc.Survey.Update(uint(id), reqForm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteSurvey 删除问卷
func (h *SurveyHandler) DeleteSurvey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	if err := h.svc.Survey.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// GetSurveyStatistics 问卷统计
func (h *SurveyHandler) GetSurveyStatistics(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	data, err := h.svc.Survey.GetStatistics(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetSurveysForStudent 员工端问卷列表
func (h *SurveyHandler) GetSurveysForStudent(c *gin.Context) {
	ident := GetCurrentIdentity(c)
	if ident == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录"})
		return
	}
	data, err := h.svc.Survey.GetStudentSurveys(ident.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetSurveyDetailForStudent 员工端问卷详情
func (h *SurveyHandler) GetSurveyDetailForStudent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	data, err := h.svc.Survey.GetDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "问卷不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// SubmitSurvey 员工提交问卷
func (h *SurveyHandler) SubmitSurvey(c *gin.Context) {
	ident := GetCurrentIdentity(c)
	if ident == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录"})
		return
	}
	var reqForm req.SubmitSurveyReq
	if err := c.ShouldBindJSON(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := h.svc.Survey.Submit(ident.Name, reqForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "提交成功"})
}

// 状态自动更新（Go 层面比较，兼容 SQLite/MySQL 时区差异）
func (h *SurveyHandler) updateSurveyStatus() {
	now := time.Now()
	loc := now.Location()
	db := h.getDB()

	var surveys []models.Survey
	db.Find(&surveys)
	for _, s := range surveys {
		startTime := time.Date(s.StartTime.Year(), s.StartTime.Month(), s.StartTime.Day(),
			s.StartTime.Hour(), s.StartTime.Minute(), s.StartTime.Second(), 0, loc)
		endTime := time.Date(s.EndTime.Year(), s.EndTime.Month(), s.EndTime.Day(),
			s.EndTime.Hour(), s.EndTime.Minute(), s.EndTime.Second(), 0, loc)

		var newStatus string
		if now.After(startTime) && now.Before(endTime) {
			newStatus = "active"
		} else if now.After(endTime) {
			newStatus = "ended"
		} else {
			newStatus = "upcoming"
		}
		if s.Status != newStatus {
			db.Model(&s).Update("status", newStatus)
		}
	}
}

// ExportSurveyDetail 导出单个问卷数据
func (h *SurveyHandler) ExportSurveyDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	format := strings.ToLower(c.DefaultQuery("format", "xlsx"))
	db := h.getDB()

	var survey models.Survey
	if err := db.First(&survey, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "问卷不存在"})
		return
	}

	// 查询该问卷的答案
	type SurveyExportRow struct {
		SurveyTitle   string
		StudentName   string
		QuestionTitle string
		QuestionType  string
		Answer        string
		SubmitTime    string
	}
	var rows []SurveyExportRow
	var answers []models.SurveyAnswer
	db.Where("survey_id = ?", survey.ID).Order("student_name, created_at").Find(&answers)

	questionMap := make(map[uint]string)
	questionTypeMap := make(map[uint]string)
	optContentMap := make(map[uint]string)

	for _, ans := range answers {
		if _, ok := questionMap[ans.SurveyQuestionID]; !ok {
			var q models.SurveyQuestion
			if db.First(&q, ans.SurveyQuestionID).Error == nil {
				questionMap[ans.SurveyQuestionID] = q.Title
				questionTypeMap[ans.SurveyQuestionID] = q.Type
			}
		}
	}

	var allOpts []models.SurveyOption
	db.Where("survey_id = ?", survey.ID).Find(&allOpts)
	for _, opt := range allOpts {
		optContentMap[opt.ID] = opt.Content
	}

	for _, ans := range answers {
		var answerText string
		answerJSON := ans.Answer

		var optIDs []uint
		if err := json.Unmarshal([]byte(answerJSON), &optIDs); err == nil && len(optIDs) > 0 {
			var labels []string
			for _, oid := range optIDs {
				if content, ok := optContentMap[oid]; ok && content != "" {
					labels = append(labels, content)
				} else {
					labels = append(labels, fmt.Sprintf("#%d", oid))
				}
			}
			answerText = strings.Join(labels, "、")
		} else {
			var strAnswers []string
			if err := json.Unmarshal([]byte(answerJSON), &strAnswers); err == nil {
				answerText = strings.Join(strAnswers, "、")
			} else {
				answerText = answerJSON
			}
		}

		rows = append(rows, SurveyExportRow{
			SurveyTitle:   survey.Title,
			StudentName:   ans.StudentName,
			QuestionTitle: questionMap[ans.SurveyQuestionID],
			QuestionType:  questionTypeMap[ans.SurveyQuestionID],
			Answer:        answerText,
			SubmitTime:    ans.CreatedAt.Local().Format("2006-01-02 15:04:05"),
		})
	}

	typeLabel := map[string]string{"single": "单选", "multiple": "多选", "fill": "填空", "essay": "简答"}
	headers := []string{"问卷标题", "填写人", "题目内容", "题目类型", "答案", "提交时间"}
	safeTitle := strings.ReplaceAll(survey.Title, "/", "_")

	if format == "csv" {
		filename := fmt.Sprintf("问卷_%s_%s.csv", safeTitle, time.Now().Format("20060102_150405"))
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
		writer := csv.NewWriter(c.Writer)
		writer.Write(headers)
		for _, row := range rows {
			writer.Write([]string{row.SurveyTitle, row.StudentName, row.QuestionTitle, typeLabel[row.QuestionType], row.Answer, row.SubmitTime})
		}
		writer.Flush()
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "问卷数据"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"059669"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	for i, hdr := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hdr)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheet, 1, 25)

	for i, row := range rows {
		rowNum := i + 2
		values := []interface{}{row.SurveyTitle, row.StudentName, row.QuestionTitle, typeLabel[row.QuestionType], row.Answer, row.SubmitTime}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowNum)
			f.SetCellValue(sheet, cell, v)
		}
	}
	f.SetColWidth(sheet, "A", "A", 30)
	f.SetColWidth(sheet, "B", "B", 16)
	f.SetColWidth(sheet, "C", "C", 40)
	f.SetColWidth(sheet, "D", "D", 10)
	f.SetColWidth(sheet, "E", "E", 30)
	f.SetColWidth(sheet, "F", "F", 20)

	filename := fmt.Sprintf("问卷_%s_%s.xlsx", safeTitle, time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

// ExportSurveys 导出问卷数据
func (h *SurveyHandler) ExportSurveys(c *gin.Context) {
	format := strings.ToLower(c.DefaultQuery("format", "xlsx"))
	db := h.getDB()

	// 获取所有问卷答案，按人分组
	type SurveyExportRow struct {
		SurveyTitle   string
		StudentName   string
		QuestionTitle string
		QuestionType  string
		Answer        string
		SubmitTime    string
	}

	var rows []SurveyExportRow

	var answers []models.SurveyAnswer
	db.Order("survey_id, student_name, created_at").Find(&answers)

	// 构建映射
	surveyTitleMap := make(map[uint]string)
	questionMap := make(map[uint]string)    // questionID → title
	questionTypeMap := make(map[uint]string) // questionID → type
	optContentMap := make(map[uint]string)   // optionID → content

	for _, ans := range answers {
		if _, ok := surveyTitleMap[ans.SurveyID]; !ok {
			var s models.Survey
			if db.Select("title").First(&s, ans.SurveyID).Error == nil {
				surveyTitleMap[ans.SurveyID] = s.Title
			}
		}
		if _, ok := questionMap[ans.SurveyQuestionID]; !ok {
			var q models.SurveyQuestion
			if db.First(&q, ans.SurveyQuestionID).Error == nil {
				questionMap[ans.SurveyQuestionID] = q.Title
				questionTypeMap[ans.SurveyQuestionID] = q.Type
			}
		}
	}

	// 收集所有问卷的所有选项以构建映射
	var allOpts []models.SurveyOption
	db.Find(&allOpts)
	for _, opt := range allOpts {
		optContentMap[opt.ID] = opt.Content
	}

	// 按 surveyID + questionID 收集选项映射
	var allQuestions []models.SurveyQuestion
	db.Find(&allQuestions)
	qOptMap := make(map[uint][]models.SurveyOption) // questionID → options
	for _, q := range allQuestions {
		for _, opt := range allOpts {
			if opt.SurveyQuestionID == q.ID {
				qOptMap[q.ID] = append(qOptMap[q.ID], opt)
			}
		}
	}

	for _, ans := range answers {
		var answerText string
		answerJSON := ans.Answer

		// 尝试解析为选项ID数组
		var optIDs []uint
		if err := json.Unmarshal([]byte(answerJSON), &optIDs); err == nil && len(optIDs) > 0 {
			var labels []string
			for _, oid := range optIDs {
				if content, ok := optContentMap[oid]; ok && content != "" {
					labels = append(labels, content)
				} else {
					labels = append(labels, fmt.Sprintf("#%d", oid))
				}
			}
			answerText = strings.Join(labels, "、")
		} else {
			// 尝试解析为字符串数组
			var strAnswers []string
			if err := json.Unmarshal([]byte(answerJSON), &strAnswers); err == nil {
				answerText = strings.Join(strAnswers, "、")
			} else {
				answerText = answerJSON
			}
		}

		rows = append(rows, SurveyExportRow{
			SurveyTitle:   surveyTitleMap[ans.SurveyID],
			StudentName:   ans.StudentName,
			QuestionTitle: questionMap[ans.SurveyQuestionID],
			QuestionType:  questionTypeMap[ans.SurveyQuestionID],
			Answer:        answerText,
			SubmitTime:    ans.CreatedAt.Local().Format("2006-01-02 15:04:05"),
		})
	}

	typeLabel := map[string]string{"single": "单选", "multiple": "多选", "fill": "填空", "essay": "简答"}
	headers := []string{"问卷标题", "填写人", "题目内容", "题目类型", "答案", "提交时间"}

	if format == "csv" {
		filename := fmt.Sprintf("问卷数据_%s.csv", time.Now().Format("20060102_150405"))
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
		writer := csv.NewWriter(c.Writer)
		writer.Write(headers)
		for _, row := range rows {
			writer.Write([]string{row.SurveyTitle, row.StudentName, row.QuestionTitle, typeLabel[row.QuestionType], row.Answer, row.SubmitTime})
		}
		writer.Flush()
		return
	}

	// Excel 导出
	f := excelize.NewFile()
	defer f.Close()
	sheet := "问卷数据"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"059669"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	for i, hdr := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hdr)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheet, 1, 25)

	for i, row := range rows {
		rowNum := i + 2
		values := []interface{}{row.SurveyTitle, row.StudentName, row.QuestionTitle, typeLabel[row.QuestionType], row.Answer, row.SubmitTime}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowNum)
			f.SetCellValue(sheet, cell, v)
		}
	}
	f.SetColWidth(sheet, "A", "A", 30)
	f.SetColWidth(sheet, "B", "B", 16)
	f.SetColWidth(sheet, "C", "C", 40)
	f.SetColWidth(sheet, "D", "D", 10)
	f.SetColWidth(sheet, "E", "E", 30)
	f.SetColWidth(sheet, "F", "F", 20)

	filename := fmt.Sprintf("问卷数据_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}
