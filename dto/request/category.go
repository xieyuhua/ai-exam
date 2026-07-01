package request

// CreateCategoryReq 创建分类请求
type CreateCategoryReq struct {
	Name string `json:"name" binding:"required"`
	Desc string `json:"desc"`
}

// UpdateCategoryReq 更新分类请求
type UpdateCategoryReq struct {
	Name string `json:"name" binding:"required"`
	Desc string `json:"desc"`
}
