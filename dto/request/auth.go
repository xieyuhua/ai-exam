package request

// LoginReq 登录请求
type LoginReq struct {
	WorkNo   string `json:"workNo" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

// ChangePasswordReq 修改密码请求
type ChangePasswordReq struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// WxAuthReq 企业微信授权回调
type WxAuthReq struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}
