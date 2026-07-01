package request

// StudentImportReq 管理员导入员工请求
type StudentImportReq struct {
	WorkNo   string `json:"workNo" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// StudentImportBatchReq 批量导入员工请求
type StudentImportBatchReq struct {
	Students []StudentImportReq `json:"students" binding:"required"`
}
