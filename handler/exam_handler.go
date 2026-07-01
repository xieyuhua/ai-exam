package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

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

// ==================== ExamHandler 考试管理 ====================

type ExamHandler struct {
	baseHandler
}

func NewExamHandler(sr *service.ServiceRegistry, db *gorm.DB) *ExamHandler {
	return &ExamHandler{baseHandler: baseHandler{svc: sr, db: db}}
}

// ==================== 考试 CRUD ====================

// GetExams 获取考试列表
// @Summary      获取考试列表
// @Description  分页查询考试，支持按分类、关键词筛选
// @Tags         考试管理
// @Accept       json
// @Produce      json
// @Param        categoryId  query  string  false  "分类 ID"
// @Param        keyword     query  string  false  "搜索关键词"
// @Param        page        query  int     false  "页码 (默认 1)"
// @Param        pageSize    query  int     false  "每页条数 (默认 10)"
// @Success      200         {object}  resp.APIResponse{data=resp.PagedData}  "考试列表"
// @Failure      500         {object}  resp.APIResponse  "查询失败"
// @Security     BearerToken
// @Router       /api/admin/exams [get]
func (h *ExamHandler) GetExams(c *gin.Context) {
	page, pageSize := util.ParsePagination(c.Query("page"), c.Query("pageSize"))
	data, err := h.svc.Exam.List(c.Query("categoryId"), c.Query("keyword"), c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// CreateExam 创建考试
// @Summary      创建考试
// @Description  创建一场新考试，可同时指定考题列表
// @Tags         考试管理
// @Accept       json
// @Produce      json
// @Param        body  body  req.CreateExamReq  true  "考试信息"
// @Success      200   {object}  resp.APIResponse{data=models.Exam}  "创建成功"
// @Failure      400   {object}  resp.APIResponse  "参数错误"
// @Failure      500   {object}  resp.APIResponse  "创建失败"
// @Security     BearerToken
// @Router       /api/admin/exams [post]
func (h *ExamHandler) CreateExam(c *gin.Context) {
	var r req.CreateExamReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	exam, err := h.svc.Exam.Create(r, h.getDB())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": exam, "message": "创建成功"})
}

// UpdateExam 更新考试
// @Summary      更新考试
// @Description  修改考试基本信息（标题、时间、时长、设置等）
// @Tags         考试管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                  true  "考试 ID"
// @Param        body  body  req.UpdateExamReq  true  "新考试信息"
// @Success      200   {object}  resp.APIResponse  "更新成功"
// @Failure      400   {object}  resp.APIResponse  "参数错误"
// @Failure      500   {object}  resp.APIResponse  "更新失败"
// @Security     BearerToken
// @Router       /api/admin/exams/{id} [put]
func (h *ExamHandler) UpdateExam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	var r req.UpdateExamReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.svc.Exam.Update(uint(id), r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteExam 删除考试
// @Summary      删除考试
// @Description  删除指定考试及其关联数据
// @Tags         考试管理
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "考试 ID"
// @Success      200  {object}  resp.APIResponse  "删除成功"
// @Failure      400  {object}  resp.APIResponse  "无效的ID"
// @Failure      500  {object}  resp.APIResponse  "删除失败"
// @Security     BearerToken
// @Router       /api/admin/exams/{id} [delete]
func (h *ExamHandler) DeleteExam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	if err := h.svc.Exam.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ==================== 考试内考题管理 ====================

// GetExamQuestionsDetail 查看考试内考题详情
// @Summary      查看考试内考题详情
// @Description  获取某场考试下所有考题及分值（含正确答案——管理端可见）
// @Tags         考试管理
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "考试 ID"
// @Success      200  {object}  resp.APIResponse{data=resp.ExamDetailData}  "考题详情"
// @Failure      404  {object}  resp.APIResponse  "考试不存在"
// @Security     BearerToken
// @Router       /api/admin/exams/{id}/questions [get]
func (h *ExamHandler) GetExamQuestionsDetail(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的考试ID"})
		return
	}

	db := h.getDB()
	uid := uint(examID)

	var exam models.Exam
	if db.First(&exam, uid).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "考试不存在"})
		return
	}

	var eqs []models.ExamQuestion
	db.Where("exam_id = ?", uid).Order("question_id asc").Find(&eqs)

	if len(eqs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{
			"exam":          exam,
			"questions":     []interface{}{},
			"totalScore":    0,
			"questionCount": 0,
		}})
		return
	}

	questionIDs := make([]uint, len(eqs))
	scoreMap := make(map[uint]float64)
	for i, eq := range eqs {
		questionIDs[i] = eq.QuestionID
		scoreMap[eq.QuestionID] = eq.Score
	}

	var questions []models.Question
	db.Where("id IN ?", questionIDs).Find(&questions)

	qMap := make(map[uint]models.Question)
	for _, q := range questions {
		qMap[q.ID] = q
	}

	list := make([]resp.ExamQuestionDetail, 0)
	totalScore := 0.0
	for _, qid := range questionIDs {
		q, ok := qMap[qid]
		if !ok {
			continue
		}
		options := service.ParseOptions(q.Options)
		answer := service.ParseAnswer(q.Answer)
		sc := scoreMap[qid]
		totalScore += sc
		list = append(list, resp.ExamQuestionDetail{
			QuestionID:  q.ID,
			CategoryID:  q.CategoryID,
			Type:        q.Type,
			Title:       q.Title,
			Options:     options,
			Answer:      answer,
			Explanation: q.Explanation,
			Score:       sc,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{
		"exam":          exam,
		"questions":     list,
		"totalScore":    totalScore,
		"questionCount": len(list),
	}})
}

// AddExamQuestions 向考试添加考题
// @Summary      向考试添加考题
// @Description  从题库选取题目加入考试，可批量添加并设置每题分值
// @Tags         考试管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                          true  "考试 ID"
// @Param        body  body  req.ImportExamQuestionsReq  true  "题目 ID 列表及分值"
// @Success      200   {object}  resp.APIResponse{data=resp.BatchImportResult}  "添加结果"
// @Failure      404   {object}  resp.APIResponse  "考试不存在"
// @Failure      500   {object}  resp.APIResponse  "添加失败"
// @Security     BearerToken
// @Router       /api/admin/exams/{id}/questions [post]
func (h *ExamHandler) AddExamQuestions(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的考试ID"})
		return
	}

	var r req.ImportExamQuestionsReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if _, err := h.svc.Exam.GetByID(uint(examID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "考试不存在"})
		return
	}

	added, skipped, err := h.svc.Exam.AddExamQuestions(uint(examID), r.QuestionIDs, r.DefaultScore, r.Scores)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "添加失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功添加 " + strconv.Itoa(added) + " 道考题", "data": gin.H{"added": added, "skipped": skipped}})
}

// UpdateExamQuestionScore 修改考题分值
// @Summary      修改考题分值
// @Description  修改某场考试中某道题目的分值
// @Tags         考试管理
// @Accept       json
// @Produce      json
// @Param        id    path  int     true  "考试 ID"
// @Param        qid   path  int     true  "题目 ID"
// @Param        body  body  object  true  "{\"score\":10.5}" 格式(json)
// @Success      200   {object}  resp.APIResponse  "分值更新成功"
// @Failure      400   {object}  resp.APIResponse  "参数错误"
// @Failure      404   {object}  resp.APIResponse  "该考题不在考试中"
// @Security     BearerToken
// @Router       /api/admin/exams/{id}/questions/{qid} [put]
func (h *ExamHandler) UpdateExamQuestionScore(c *gin.Context) {
	examID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	qid, _ := strconv.ParseUint(c.Param("qid"), 10, 64)

	var r struct {
		Score float64 `json:"score"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	rows, err := h.svc.Exam.UpdateExamQuestionScore(uint(examID), uint(qid), r.Score)
	if err != nil || rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "该考题不在考试中"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "分值更新成功"})
}

// RemoveExamQuestion 从考试中移除考题
// @Summary      从考试中移除考题
// @Description  从指定考试中删除一道题目
// @Tags         考试管理
// @Accept       json
// @Produce      json
// @Param        id   path  int  true  "考试 ID"
// @Param        qid  path  int  true  "题目 ID"
// @Success      200  {object}  resp.APIResponse  "移除成功"
// @Failure      404  {object}  resp.APIResponse  "该考题不在考试中"
// @Security     BearerToken
// @Router       /api/admin/exams/{id}/questions/{qid} [delete]
func (h *ExamHandler) RemoveExamQuestion(c *gin.Context) {
	examID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	qid, _ := strconv.ParseUint(c.Param("qid"), 10, 64)

	if err := h.svc.Exam.RemoveExamQuestion(uint(examID), uint(qid)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "该考题不在考试中"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "移除成功"})
}

// ExportExamQuestions 导出考题到 Excel（按导入模板格式）
// @Summary      导出考题
// @Description  导出考试内所有考题为 Excel 文件，格式与导入模板一致，方便修改后重新导入覆盖
// @Tags         考试管理
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        id  path  int  true  "考试 ID"
// @Success      200  {file}  binary  "Excel 文件"
// @Failure      404  {object}  resp.APIResponse  "考试不存在"
// @Security     BearerToken
// @Router       /api/admin/exams/{id}/questions/export [get]
func (h *ExamHandler) ExportExamQuestions(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的考试ID"})
		return
	}

	db := h.getDB()
	uid := uint(examID)

	var exam models.Exam
	if db.First(&exam, uid).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "考试不存在"})
		return
	}

	// 获取考试内所有考题关联
	var eqs []models.ExamQuestion
	db.Where("exam_id = ?", uid).Order("question_id asc").Find(&eqs)
	if len(eqs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "该考试暂无考题可导出"})
		return
	}

	questionIDs := make([]uint, len(eqs))
	scoreMap := make(map[uint]float64, len(eqs))
	for i, eq := range eqs {
		questionIDs[i] = eq.QuestionID
		scoreMap[eq.QuestionID] = eq.Score
	}

	var questions []models.Question
	db.Where("id IN ?", questionIDs).Find(&questions)
	qMap := make(map[uint]models.Question)
	for _, q := range questions {
		qMap[q.ID] = q
	}

	// 生成 Excel
	f := excelize.NewFile()
	defer f.Close()

	sheet := "考题数据"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建 Excel 工作表失败"})
		return
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"667eea"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	headers := []string{"题型", "题目内容", "A", "B", "C", "D", "正确答案", "分值", "答案解析"}
	for i, hdr := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hdr)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// 题型中文映射
	typeName := map[string]string{
		"single": "单选题", "multiple": "多选题", "judge": "判断题",
		"fill": "填空题", "essay": "简答题",
	}

	rowIdx := 2
	for _, qid := range questionIDs {
		q, ok := qMap[qid]
		if !ok {
			continue
		}
		// 解析选项和答案
		opts := service.ParseOptions(q.Options)
		answer := service.ParseAnswer(q.Answer)

		// 列1：题型
		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIdx), typeName[q.Type])

		// 列2：题目内容
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIdx), q.Title)

		// 列3-6：选项 A/B/C/D（按 label 映射）
		optMap := make(map[string]string)
		for _, o := range opts {
			optMap[strings.ToUpper(o.Label)] = o.Content
		}
		for _, label := range []string{"A", "B", "C", "D"} {
			col, _ := excelize.CoordinatesToCellName(int(label[0]-'A')+3, rowIdx)
			f.SetCellValue(sheet, col, optMap[label])
		}

		// 列7：正确答案
		f.SetCellValue(sheet, fmt.Sprintf("G%d", rowIdx), strings.Join(answer, ","))

		// 列8：分值
		f.SetCellValue(sheet, fmt.Sprintf("H%d", rowIdx), scoreMap[qid])

		// 列9：答案解析
		f.SetCellValue(sheet, fmt.Sprintf("I%d", rowIdx), q.Explanation)

		rowIdx++
	}

	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", "B", 40)
	f.SetColWidth(sheet, "C", "I", 16)

	filename := fmt.Sprintf("考题导出_%s.xlsx", exam.Title)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", urlEncode(filename)))
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "导出 Excel 失败"})
		return
	}
}

// ImportExamQuestions 表格导入考题到考试
// @Summary      表格导入考题到考试
// @Description  上传 Excel/CSV 文件，自动创建题目到题库并加入指定考试
// @Tags         考试管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path  int   true  "考试 ID"
// @Param        file formData  file  true  "导入文件 (.xlsx/.csv/.json)"
// @Success      200  {object}  resp.APIResponse  "导入成功"
// @Failure      400  {object}  resp.APIResponse  "参数错误或文件格式错误"
// @Failure      404  {object}  resp.APIResponse  "考试不存在"
// @Failure      500  {object}  resp.APIResponse  "导入失败"
// @Security     BearerToken
// @Router       /api/admin/exams/{id}/questions/import [post]
func (h *ExamHandler) ImportExamQuestions(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的考试ID"})
		return
	}

	db := h.getDB()
	uid := uint(examID)

	var exam models.Exam
	if db.First(&exam, uid).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "考试不存在"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件读取失败"})
		return
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	var importData []service.ImportQuestionRow

	switch ext {
	case ".json":
		importData, err = parseJSONImport(src)
	case ".csv":
		importData, err = parseCSVImport(src)
	case ".xlsx":
		importData, err = parseXLSXImport(src)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不支持的文件格式，请上传 .xlsx/.csv/.json"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("文件解析失败: %s", err.Error())})
		return
	}

	if len(importData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "文件中未解析到任何题目数据"})
		return
	}

	// 批量创建/更新题目到题库 + 关联到考试
	added := 0
	updated := 0
	skipped := 0
	for _, item := range importData {
		opts := service.ConvertOptions(item.Options)
		optionsJSON, _ := json.Marshal(opts)
		answerJSON, _ := json.Marshal(item.Answer)

		// 先按标题+分类查找是否已存在相同题目
		var existing models.Question
		err := db.Where("category_id = ? AND title = ?", exam.CategoryID, item.Title).First(&existing).Error

		var qid uint
		if err == nil {
			// 已存在相同标题的题目 → 覆盖更新
			db.Model(&existing).Updates(map[string]interface{}{
				"type":        item.Type,
				"options":     string(optionsJSON),
				"answer":      string(answerJSON),
				"explanation": item.Explanation,
			})
			qid = existing.ID
			updated++
		} else {
			// 不存在 → 新建
			q := models.Question{
				CategoryID:  exam.CategoryID,
				Type:        item.Type,
				Title:       item.Title,
				Options:     string(optionsJSON),
				Answer:      string(answerJSON),
				Explanation: item.Explanation,
			}
			if err := db.Create(&q).Error; err != nil {
				skipped++
				continue
			}
			qid = q.ID
		}

		// 关联到考试（如果尚未关联）
		var eq models.ExamQuestion
		if db.Where("exam_id = ? AND question_id = ?", uid, qid).First(&eq).Error != nil {
			score := item.Score // 导入的分值，未填则默认 0
			if err := db.Create(&models.ExamQuestion{
				ExamID:     uid,
				QuestionID: qid,
				Score:      score,
			}).Error; err != nil {
				skipped++
			} else {
				added++
			}
		} else if item.Score > 0 {
			// 已存在关联但分值>0 → 更新分值
			db.Model(&eq).Update("score", item.Score)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("成功导入 %d 道考题（更新 %d 道，跳过 %d 道）", added, updated, skipped),
		"data":    gin.H{"added": added, "updated": updated, "skipped": skipped},
	})
}

// ==================== 文件解析辅助函数（公共） ====================

func parseJSONImport(r io.Reader) ([]service.ImportQuestionRow, error) {
	var data []struct {
		Type        string            `json:"type"`
		Title       string            `json:"title"`
		Options     map[string]string `json:"options"`
		Answer      []string          `json:"answer"`
		Explanation string            `json:"explanation"`
	}
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, fmt.Errorf("JSON 格式错误: %w", err)
	}
	result := make([]service.ImportQuestionRow, len(data))
	for i, d := range data {
		result[i] = service.ImportQuestionRow{
			Type:        d.Type,
			Title:       d.Title,
			Options:     d.Options,
			Answer:      d.Answer,
			Explanation: d.Explanation,
		}
	}
	return result, nil
}

func parseCSVImport(r io.Reader) ([]service.ImportQuestionRow, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("CSV 至少需要表头和一行数据")
	}
	var headers []string
	for _, h := range strings.Split(lines[0], ",") {
		headers = append(headers, strings.TrimSpace(h))
	}
	var dataRows [][]string
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var cols []string
		for _, c := range strings.Split(line, ",") {
			cols = append(cols, strings.TrimSpace(c))
		}
		dataRows = append(dataRows, cols)
	}
	return parseTableRows(headers, dataRows), nil
}

func parseXLSXImport(r io.Reader) ([]service.ImportQuestionRow, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("Excel 解析失败: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel 至少需要表头和一行数据")
	}
	return parseTableRows(rows[0], rows[1:]), nil
}

func parseTableRows(headers []string, dataRows [][]string) []service.ImportQuestionRow {
	colMap := make(map[string]int)
	for i, hdr := range headers {
		trimmed := strings.TrimSpace(hdr)
		colMap[trimmed] = i
		colMap[strings.ToLower(trimmed)] = i
	}

	var result []service.ImportQuestionRow
	for _, row := range dataRows {
		if len(row) < len(headers) {
			padded := make([]string, len(headers))
			copy(padded, row)
			row = padded
		}
		qType := getCol(row, colMap, "type", "题型")
		title := getCol(row, colMap, "title", "题目", "题目内容")
		answerStr := getCol(row, colMap, "answer", "正确答案", "答案")
		scoreStr := getCol(row, colMap, "score", "分值")
		explanation := getCol(row, colMap, "explanation", "答案解析", "解析")

		if title == "" {
			continue
		}
		if qType == "" {
			qType = "single"
		}
		qType = util.NormalizeQuestionType(qType)

		options := make(map[string]string)
		for _, col := range headers {
			label := extractOptionLabel(col)
			if label == "" {
				continue
			}
			if idx, ok := colMap[col]; ok && idx < len(row) && row[idx] != "" {
				options[label] = row[idx]
			}
		}

		answer := parseAnswerString(answerStr)
		score, _ := strconv.ParseFloat(scoreStr, 64)

		result = append(result, service.ImportQuestionRow{
			Type:        qType,
			Title:       title,
			Options:     options,
			Answer:      answer,
			Explanation: explanation,
			Score:       score,
		})
	}
	return result
}

func getCol(row []string, colMap map[string]int, keys ...string) string {
	for _, k := range keys {
		if idx, ok := colMap[k]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
	}
	return ""
}

func extractOptionLabel(col string) string {
	lower := strings.ToLower(strings.TrimSpace(col))
	// 支持 "选项A" / "optionA" 格式
	for _, prefix := range []string{"选项", "option"} {
		after, found := strings.CutPrefix(lower, prefix)
		if found {
			after = strings.TrimSpace(after)
			after = strings.TrimPrefix(after, "_")
			if len(after) == 1 && after[0] >= 'a' && after[0] <= 'h' {
				return after
			}
		}
	}
	// 支持纯字母列头：A/B/C/D 等（模板导出格式）
	if len(lower) == 1 && lower[0] >= 'a' && lower[0] <= 'h' {
		return lower
	}
	return ""
}

func parseAnswerString(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}
	separators := []string{",", "、", "，", ";", "；", "|", "/"}
	for _, sep := range separators {
		s = strings.ReplaceAll(s, sep, ",")
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, strings.ToLower(p))
		}
	}
	return result
}

// urlEncode 对文件名做 URL 编码（兼容中文文件名）
func urlEncode(s string) string {
	return url.QueryEscape(s)
}

// GetAvailableQuestions 获取可导入的题目
// @Summary      获取可导入的题目列表
// @Description  查询同分类下尚未加入指定考试的题目，供管理员选题导入
// @Tags         考试管理
// @Accept       json
// @Produce      json
// @Param        examId   query  int     true   "考试 ID"
// @Param        keyword  query  string  false  "搜索关键词"
// @Param        type     query  string  false  "题型筛选"
// @Success      200      {object}  resp.APIResponse  "可导入题目列表"
// @Failure      400      {object}  resp.APIResponse  "缺少 examId"
// @Failure      404      {object}  resp.APIResponse  "考试不存在"
// @Security     BearerToken
// @Router       /api/admin/questions/available [get]
func (h *ExamHandler) GetAvailableQuestions(c *gin.Context) {
	examIDStr := c.Query("examId")
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

	var eqs []models.ExamQuestion
	db.Where("exam_id = ?", uid).Find(&eqs)
	existingIDs := make([]uint, 0, len(eqs))
	for _, eq := range eqs {
		existingIDs = append(existingIDs, eq.QuestionID)
	}

	keyword := c.Query("keyword")
	qType := c.Query("type")
	questions, err := h.svc.Exam.GetAvailableQuestions(exam.CategoryID, existingIDs, keyword, qType, 100)
	if err != nil {
		// fallback
		query := db.Model(&models.Question{}).Where("category_id = ?", exam.CategoryID)
		if len(existingIDs) > 0 {
			query = query.Where("id NOT IN ?", existingIDs)
		}
		if keyword != "" {
			query = query.Where("title LIKE ?", "%"+keyword+"%")
		}
		if qType != "" {
			query = query.Where("type = ?", qType)
		}
		query.Order("id desc").Limit(100).Find(&questions)
	}

	type availQ struct {
		ID      uint            `json:"id"`
		Type    string          `json:"type"`
		Title   string          `json:"title"`
		Options []dto.OptionPair `json:"options"`
	}
	result := make([]availQ, len(questions))
	for i, q := range questions {
		result[i] = availQ{
			ID:      q.ID,
			Type:    q.Type,
			Title:   q.Title,
			Options: service.ParseOptions(q.Options),
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}
