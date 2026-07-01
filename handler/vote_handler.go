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

type VoteHandler struct {
	baseHandler
}

func NewVoteHandler(sr *service.ServiceRegistry, db *gorm.DB) *VoteHandler {
	return &VoteHandler{baseHandler: baseHandler{svc: sr, db: db}}
}

// GetVotes 获取投票列表
func (h *VoteHandler) GetVotes(c *gin.Context) {
	page, pageSize := util.ParsePagination(c.Query("page"), c.Query("pageSize"))
	data, err := h.svc.Vote.List(c.Query("keyword"), c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// CreateVote 创建投票
func (h *VoteHandler) CreateVote(c *gin.Context) {
	var reqForm req.CreateVoteReq
	if err := c.ShouldBindJSON(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	v, err := h.svc.Vote.Create(reqForm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": v, "message": "创建成功"})
}

// GetVoteDetail 获取投票详情
func (h *VoteHandler) GetVoteDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	data, err := h.svc.Vote.GetDetail(uint(id), "")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "投票不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// UpdateVote 更新投票
func (h *VoteHandler) UpdateVote(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	var reqForm req.CreateVoteReq
	if err := c.ShouldBindJSON(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := h.svc.Vote.Update(uint(id), reqForm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteVote 删除投票
func (h *VoteHandler) DeleteVote(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	if err := h.svc.Vote.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// GetVotesForStudent 员工端投票列表
func (h *VoteHandler) GetVotesForStudent(c *gin.Context) {
	ident := GetCurrentIdentity(c)
	if ident == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录"})
		return
	}
	data, err := h.svc.Vote.GetStudentVotes(ident.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetVoteDetailForStudent 员工端投票详情
func (h *VoteHandler) GetVoteDetailForStudent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	ident := GetCurrentIdentity(c)
	if ident == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录"})
		return
	}
	data, err := h.svc.Vote.GetDetail(uint(id), ident.Name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "投票不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// SubmitVote 员工提交投票
func (h *VoteHandler) SubmitVote(c *gin.Context) {
	ident := GetCurrentIdentity(c)
	if ident == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录"})
		return
	}
	var reqForm req.SubmitVoteReq
	if err := c.ShouldBindJSON(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := h.svc.Vote.Submit(ident.Name, reqForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "投票成功"})
}

// ExportVoteDetail 导出单个投票数据
func (h *VoteHandler) ExportVoteDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	format := strings.ToLower(c.DefaultQuery("format", "xlsx"))
	db := h.getDB()

	// 查询该投票
	var vote models.Vote
	if err := db.First(&vote, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "投票不存在"})
		return
	}

	// 查询该投票的选项
	var opts []models.VoteOption
	db.Where("vote_id = ?", vote.ID).Find(&opts)
	optContentMap := make(map[uint]string)
	for _, opt := range opts {
		optContentMap[opt.ID] = opt.Content
	}

	// 查询该投票的记录
	type VoteExportRow struct {
		VoteTitle   string
		StudentName string
		Options     string
		VoteTime    string
	}
	var rows []VoteExportRow
	var records []models.VoteRecord
	db.Where("vote_id = ?", vote.ID).Order("id desc").Find(&records)
	for _, rec := range records {
		var oids []uint
		json.Unmarshal([]byte(rec.OptionIDs), &oids)
		var optLabels []string
		for _, oid := range oids {
			if content, ok := optContentMap[oid]; ok && content != "" {
				optLabels = append(optLabels, content)
			} else {
				optLabels = append(optLabels, fmt.Sprintf("#%d", oid))
			}
		}
		rows = append(rows, VoteExportRow{
			VoteTitle:   vote.Title,
			StudentName: rec.StudentName,
			Options:     strings.Join(optLabels, "、"),
			VoteTime:    rec.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	headers := []string{"投票标题", "投票人", "选择的选项", "投票时间"}
	safeTitle := strings.ReplaceAll(vote.Title, "/", "_")

	if format == "csv" {
		filename := fmt.Sprintf("投票_%s_%s.csv", safeTitle, time.Now().Format("20060102_150405"))
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
		writer := csv.NewWriter(c.Writer)
		writer.Write(headers)
		for _, row := range rows {
			writer.Write([]string{row.VoteTitle, row.StudentName, row.Options, row.VoteTime})
		}
		writer.Flush()
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "投票数据"
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

	for i, row := range rows {
		rowNum := i + 2
		values := []interface{}{row.VoteTitle, row.StudentName, row.Options, row.VoteTime}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowNum)
			f.SetCellValue(sheet, cell, v)
		}
	}
	f.SetColWidth(sheet, "A", "A", 30)
	f.SetColWidth(sheet, "B", "B", 16)
	f.SetColWidth(sheet, "C", "C", 40)
	f.SetColWidth(sheet, "D", "D", 20)

	filename := fmt.Sprintf("投票_%s_%s.xlsx", safeTitle, time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

// ExportVotes 导出投票数据
func (h *VoteHandler) ExportVotes(c *gin.Context) {
	format := strings.ToLower(c.DefaultQuery("format", "xlsx"))
	db := h.getDB()

	// 查询所有投票记录
	type VoteExportRow struct {
		VoteTitle   string
		StudentName string
		Options     string
		VoteTime    string
	}

	var rows []VoteExportRow

	// 获取所有投票记录
	var records []models.VoteRecord
	db.Order("id desc").Find(&records)

	// 构建 voteID→title 映射
	voteTitleMap := make(map[uint]string)
	for _, rec := range records {
		if _, ok := voteTitleMap[rec.VoteID]; !ok {
			var v models.Vote
			if db.Select("title").First(&v, rec.VoteID).Error == nil {
				voteTitleMap[rec.VoteID] = v.Title
			}
		}
	}

	// 构建 optionID→content 映射
	optContentMap := make(map[uint]string)
	var allOpts []models.VoteOption
	db.Find(&allOpts)
	for _, opt := range allOpts {
		optContentMap[opt.ID] = opt.Content
	}

	for _, rec := range records {
		var oids []uint
		json.Unmarshal([]byte(rec.OptionIDs), &oids)
		var optLabels []string
		for _, oid := range oids {
			if content, ok := optContentMap[oid]; ok && content != "" {
				optLabels = append(optLabels, content)
			} else {
				optLabels = append(optLabels, fmt.Sprintf("#%d", oid))
			}
		}
		rows = append(rows, VoteExportRow{
			VoteTitle:   voteTitleMap[rec.VoteID],
			StudentName: rec.StudentName,
			Options:     strings.Join(optLabels, "、"),
			VoteTime:    rec.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	headers := []string{"投票标题", "投票人", "选择的选项", "投票时间"}

	if format == "csv" {
		filename := fmt.Sprintf("投票数据_%s.csv", time.Now().Format("20060102_150405"))
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM
		writer := csv.NewWriter(c.Writer)
		writer.Write(headers)
		for _, row := range rows {
			writer.Write([]string{row.VoteTitle, row.StudentName, row.Options, row.VoteTime})
		}
		writer.Flush()
		return
	}

	// Excel 导出
	f := excelize.NewFile()
	defer f.Close()
	sheet := "投票数据"
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

	for i, row := range rows {
		rowNum := i + 2
		values := []interface{}{row.VoteTitle, row.StudentName, row.Options, row.VoteTime}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowNum)
			f.SetCellValue(sheet, cell, v)
		}
	}
	f.SetColWidth(sheet, "A", "A", 30)
	f.SetColWidth(sheet, "B", "B", 16)
	f.SetColWidth(sheet, "C", "C", 40)
	f.SetColWidth(sheet, "D", "D", 20)

	filename := fmt.Sprintf("投票数据_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

// 状态自动更新（Go 层面比较，兼容 SQLite/MySQL 时区差异）
func (h *VoteHandler) updateVoteStatus() {
	now := time.Now()
	loc := now.Location()
	db := h.getDB()

	var votes []models.Vote
	db.Find(&votes)
	for _, v := range votes {
		startTime := time.Date(v.StartTime.Year(), v.StartTime.Month(), v.StartTime.Day(),
			v.StartTime.Hour(), v.StartTime.Minute(), v.StartTime.Second(), 0, loc)
		endTime := time.Date(v.EndTime.Year(), v.EndTime.Month(), v.EndTime.Day(),
			v.EndTime.Hour(), v.EndTime.Minute(), v.EndTime.Second(), 0, loc)

		var newStatus string
		if now.After(startTime) && now.Before(endTime) {
			newStatus = "active"
		} else if now.After(endTime) {
			newStatus = "ended"
		} else {
			newStatus = "upcoming"
		}
		if v.Status != newStatus {
			db.Model(&v).Update("status", newStatus)
		}
	}
}
