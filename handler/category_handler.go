package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	req "exam-system/dto/request"
	"exam-system/service"
)

// ==================== CategoryHandler 分类管理 ====================

type CategoryHandler struct {
	baseHandler
}

func NewCategoryHandler(sr *service.ServiceRegistry) *CategoryHandler {
	return &CategoryHandler{baseHandler: baseHandler{svc: sr}}
}

// GetCategories 获取分类列表
// @Summary      获取考试分类列表
// @Description  返回所有考试分类（含各级别下的考试数和题目数）
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  APIResponse{data=[]CategoryWithCount}  "分类列表"
// @Failure      500  {object}  APIResponse  "查询失败"
// @Security     BearerToken
// @Router       /api/admin/categories [get]
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	result, err := h.svc.Category.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// CreateCategory 创建分类
// @Summary      创建考试分类
// @Description  新增一个考试分类
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Param        body  body  req.CreateCategoryReq  true  "分类信息"
// @Success      200   {object}  APIResponse{data=Category}  "创建成功"
// @Failure      400   {object}  APIResponse  "参数错误"
// @Failure      500   {object}  APIResponse  "创建失败"
// @Security     BearerToken
// @Router       /api/admin/categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var r req.CreateCategoryReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	cat, err := h.svc.Category.Create(r.Name, r.Desc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": cat, "message": "创建成功"})
}

// UpdateCategory 更新分类
// @Summary      更新考试分类
// @Description  修改指定分类的名称和描述
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                       true  "分类 ID"
// @Param        body  body  req.UpdateCategoryReq  true  "新分类信息"
// @Success      200   {object}  APIResponse  "更新成功"
// @Failure      400   {object}  APIResponse  "参数错误"
// @Failure      500   {object}  APIResponse  "更新失败"
// @Security     BearerToken
// @Router       /api/admin/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	var r req.UpdateCategoryReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.svc.Category.Update(uint(id), r.Name, r.Desc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteCategory 删除分类
// @Summary      删除考试分类
// @Description  删除指定分类（如果分类下有考试或题目则不允许删除）
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "分类 ID"
// @Success      200  {object}  APIResponse  "删除成功"
// @Failure      400  {object}  APIResponse  "删除失败（含关联数据）"
// @Security     BearerToken
// @Router       /api/admin/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	if err := h.svc.Category.Delete(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
