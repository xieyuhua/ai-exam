package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	req "exam-system/dto/request"
	"exam-system/service"
	"exam-system/util"
)

// ==================== StudentHandler 员工管理 ====================

type StudentHandler struct {
	baseHandler
}

func NewStudentHandler(sr *service.ServiceRegistry) *StudentHandler {
	return &StudentHandler{baseHandler: baseHandler{svc: sr}}
}

// GetStudents 获取员工列表
// @Summary      获取员工列表
// @Description  分页查询员工，支持关键词搜索（账号/姓名）
// @Tags         员工管理
// @Accept       json
// @Produce      json
// @Param        keyword   query  string  false  "搜索关键词"
// @Param        page      query  int     false  "页码 (默认 1)"
// @Param        pageSize  query  int     false  "每页条数 (默认 10)"
// @Success      200       {object}  APIResponse{data=PagedData}  "员工列表"
// @Failure      500       {object}  APIResponse  "查询失败"
// @Security     BearerToken
// @Router       /api/admin/students [get]
func (h *StudentHandler) GetStudents(c *gin.Context) {
	page, pageSize := util.ParsePagination(c.Query("page"), c.Query("pageSize"))
	data, err := h.svc.Student.List(c.Query("keyword"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// CreateStudent 创建员工
// @Summary      添加单个员工
// @Description  管理员手动添加一名员工（账号+姓名+初始密码）
// @Tags         员工管理
// @Accept       json
// @Produce      json
// @Param        body  body  req.StudentImportReq  true  "员工信息"
// @Success      200   {object}  APIResponse{data=Student}  "员工添加成功"
// @Failure      400   {object}  APIResponse  "参数错误或账号已存在"
// @Security     BearerToken
// @Router       /api/admin/students [post]
func (h *StudentHandler) CreateStudent(c *gin.Context) {
	var r req.StudentImportReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	student, err := h.svc.Student.Create(r.WorkNo, r.Name, r.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": student, "message": "员工添加成功"})
}

// BatchCreateStudents 批量导入员工
// @Summary      批量导入员工
// @Description  一次性导入多名员工，自动跳过已存在的账号
// @Tags         员工管理
// @Accept       json
// @Produce      json
// @Param        body  body  req.StudentImportBatchReq  true  "员工列表"
// @Success      200   {object}  APIResponse{data=BatchImportResult}  "导入结果"
// @Failure      400   {object}  APIResponse  "参数错误"
// @Security     BearerToken
// @Router       /api/admin/students/batch [post]
func (h *StudentHandler) BatchCreateStudents(c *gin.Context) {
	var r req.StudentImportBatchReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if len(r.Students) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供至少一个员工信息"})
		return
	}

	added, skipped := h.svc.Student.BatchCreate(r.Students)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": fmt.Sprintf("成功导入 %d 名员工，跳过 %d 名", added, skipped), "data": gin.H{"added": added, "skipped": skipped}})
}

// UpdateStudent 更新员工信息
// @Summary      更新员工信息
// @Description  修改指定员工的姓名和/或密码
// @Tags         员工管理
// @Accept       json
// @Produce      json
// @Param        id    path  int     true  "员工 ID"
// @Param        body  body  object  true  "{\"name\":\"string\",\"password\":\"string\"}" 格式(json)
// @Success      200   {object}  APIResponse  "员工信息已更新"
// @Failure      400   {object}  APIResponse  "参数错误"
// @Security     BearerToken
// @Router       /api/admin/students/{id} [put]
func (h *StudentHandler) UpdateStudent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	var r struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.svc.Student.Update(uint(id), r.Name, r.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "员工信息已更新"})
}

// ImportStudentsExcel 通过 Excel 文件批量导入员工
// @Summary      Excel 批量导入员工
// @Description  上传 Excel 文件（.xlsx），表头需包含：账号、姓名、密码（密码列可选，默认 123456）
// @Tags         员工管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Excel 文件"
// @Success      200   {object}  APIResponse{data=BatchImportResult}  "导入结果"
// @Failure      400   {object}  APIResponse  "文件读取失败"
// @Security     BearerToken
// @Router       /api/admin/students/import [post]
func (h *StudentHandler) ImportStudentsExcel(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传 Excel 文件"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "读取文件失败"})
		return
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Excel 解析失败，请确保是 .xlsx 格式"})
		return
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil || len(rows) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Excel 至少需要表头和一行数据"})
		return
	}

	colMap := make(map[string]int)
	for i, h := range rows[0] {
		colMap[strings.TrimSpace(h)] = i
	}

	var reqs []req.StudentImportReq
	for _, row := range rows[1:] {
		get := func(key string) string {
			if idx, ok := colMap[key]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		workNo := get("账号")
		name := get("姓名")
		password := get("密码")
		if password == "" {
			password = "123456"
		}
		if workNo == "" || name == "" {
			continue
		}
		reqs = append(reqs, req.StudentImportReq{
			WorkNo:   workNo,
			Name:     name,
			Password: password,
		})
	}

	if len(reqs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "未读取到有效员工数据，请检查 Excel 表头是否包含：账号、姓名"})
		return
	}

	added, skipped := h.svc.Student.BatchCreate(reqs)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": fmt.Sprintf("成功导入 %d 名员工，跳过 %d 名", added, skipped), "data": gin.H{"added": added, "skipped": skipped}})
}

// DeleteStudent 删除员工
// @Summary      删除员工
// @Description  删除指定员工及其相关数据
// @Tags         员工管理
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "员工 ID"
// @Success      200  {object}  APIResponse  "员工已删除"
// @Failure      400  {object}  APIResponse  "无效的ID"
// @Failure      500  {object}  APIResponse  "删除失败"
// @Security     BearerToken
// @Router       /api/admin/students/{id} [delete]
func (h *StudentHandler) DeleteStudent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	if err := h.svc.Student.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "员工已删除"})
}
