package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	req "exam-system/dto/request"
	"exam-system/service"
	"exam-system/util"
)

// ==================== QuestionHandler 题目管理 ====================

type QuestionHandler struct {
	baseHandler
}

func NewQuestionHandler(sr *service.ServiceRegistry, db *gorm.DB) *QuestionHandler {
	return &QuestionHandler{baseHandler: baseHandler{svc: sr, db: db}}
}

// GetQuestions 获取题目列表
// @Summary      获取题目列表
// @Description  分页查询题库题目，支持按分类、关键词、题型筛选
// @Tags         题目管理
// @Accept       json
// @Produce      json
// @Param        categoryId  query  string  false  "分类 ID"
// @Param        keyword     query  string  false  "搜索关键词（题干）"
// @Param        type        query  string  false  "题型：single/multiple/judge/fill/essay"
// @Param        page        query  int     false  "页码 (默认 1)"
// @Param        pageSize    query  int     false  "每页条数 (默认 10)"
// @Success      200         {object}  APIResponse{data=PagedData}  "题目列表"
// @Failure      500         {object}  APIResponse  "查询失败"
// @Security     BearerToken
// @Router       /api/admin/questions [get]
func (h *QuestionHandler) GetQuestions(c *gin.Context) {
	page, pageSize := util.ParsePagination(c.Query("page"), c.Query("pageSize"))
	data, err := h.svc.Question.List(c.Query("categoryId"), c.Query("keyword"), c.Query("type"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// CreateQuestion 创建题目
// @Summary      创建题目
// @Description  新增一道题目到题库
// @Tags         题目管理
// @Accept       json
// @Produce      json
// @Param        body  body  req.CreateQuestionReq  true  "题目信息"
// @Success      200   {object}  APIResponse{data=Question}  "添加成功"
// @Failure      400   {object}  APIResponse  "参数错误"
// @Failure      500   {object}  APIResponse  "创建失败"
// @Security     BearerToken
// @Router       /api/admin/questions [post]
func (h *QuestionHandler) CreateQuestion(c *gin.Context) {
	var r req.CreateQuestionReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	q, err := h.svc.Question.Create(r)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": q, "message": "添加成功"})
}

// UpdateQuestion 更新题目
// @Summary      更新题目
// @Description  修改指定题目的内容、选项、答案等
// @Tags         题目管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                      true  "题目 ID"
// @Param        body  body  req.UpdateQuestionReq  true  "新题目信息"
// @Success      200   {object}  APIResponse  "更新成功"
// @Failure      400   {object}  APIResponse  "参数错误"
// @Security     BearerToken
// @Router       /api/admin/questions/{id} [put]
func (h *QuestionHandler) UpdateQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	var r req.UpdateQuestionReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.svc.Question.Update(uint(id), r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteQuestion 删除题目
// @Summary      删除题目
// @Description  删除指定题目
// @Tags         题目管理
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "题目 ID"
// @Success      200  {object}  APIResponse  "删除成功"
// @Failure      400  {object}  APIResponse  "无效的ID"
// @Failure      500  {object}  APIResponse  "删除失败"
// @Security     BearerToken
// @Router       /api/admin/questions/{id} [delete]
func (h *QuestionHandler) DeleteQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	if err := h.svc.Question.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// DownloadTemplate 下载导入模板
// @Summary      下载考题导入模板
// @Description  下载 Excel 格式的导入模板文件，含表头和格式说明
// @Tags         题目管理
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success      200  {file}  binary  "Excel 模板文件"
// @Security     BearerToken
// @Router       /api/admin/template [get]
func (h *QuestionHandler) DownloadTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "考题导入模板"
	f.SetSheetName("Sheet1", sheet)

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

	samples := [][]string{
		{"单选题", "以下哪个是中国的首都？", "上海", "北京", "广州", "深圳", "B", "5", "北京是中华人民共和国的首都。"},
		{"多选题", "以下哪些是哺乳动物？", "鲸鱼", "鲨鱼", "海豚", "鳄鱼", "A,C", "10", "鲸鱼和海豚是海洋哺乳动物。"},
		{"判断题", "地球围绕太阳公转。", "正确", "错误", "", "", "A", "3", "地球沿椭圆轨道围绕太阳公转，周期约365天。"},
		{"填空题", "水的化学式是___。", "", "", "", "", "H2O", "5", "水由两个氢原子和一个氧原子组成。"},
		{"简答题", "请简述社会主义核心价值观的基本内容。", "", "", "", "", "富强、民主、文明、和谐；自由、平等、公正、法治；爱国、敬业、诚信、友善", "10", "社会主义核心价值观分为国家、社会、公民三个层面。"},
	}
	for i, row := range samples {
		for j, val := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", "B", 40)
	f.SetColWidth(sheet, "C", "I", 16)

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=考题导入模板.xlsx")
	f.Write(c.Writer)
}

// ImportQuestions 批量导入题目
// @Summary      批量导入题目
// @Description  通过上传文件（json/csv/xlsx）批量导入题目到指定分类
// @Tags         题目管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        categoryId  formData  int     true   "分类 ID"
// @Param        file        formData  file    true   "导入文件 (.json/.csv/.xlsx)"
// @Success      200         {object}  APIResponse  "导入成功"
// @Failure      400         {object}  APIResponse  "参数错误或文件格式错误"
// @Failure      500         {object}  APIResponse  "导入失败"
// @Security     BearerToken
// @Router       /api/admin/questions/import [post]
func (h *QuestionHandler) ImportQuestions(c *gin.Context) {
	categoryIDStr := c.PostForm("categoryId")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择分类"})
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
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不支持的文件格式"})
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

	count, err := h.svc.Question.ImportBatch(uint(categoryID), importData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "导入失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": fmt.Sprintf("成功导入 %d 道题目", count)})
}
