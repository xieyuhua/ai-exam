package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	req "exam-system/dto/request"
	"exam-system/service"
)

// ==================== AuthHandler 认证处理 ====================

type AuthHandler struct {
	baseHandler
}

func NewAuthHandler(sr *service.ServiceRegistry) *AuthHandler {
	return &AuthHandler{baseHandler: baseHandler{svc: sr}}
}

// GetWxWorkAuthURL 获取企业微信授权 URL
// @Summary      获取企业微信授权 URL
// @Description  构造企业微信 OAuth2 授权链接，用户访问后跳转到企微授权页
// @Tags         认证
// @Accept       json
// @Produce      json
// @Success      200  {object}  APIResponse{data=WxWorkAuthData}  "成功返回授权 URL"
// @Failure      400  {object}  APIResponse  "请求失败"
// @Router       /api/auth/wxwork/url [get]
func (h *AuthHandler) GetWxWorkAuthURL(c *gin.Context) {
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	authURL, err := service.GetWxWorkAuthURL(c.Request.Host, scheme)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"authUrl": authURL}})
}

// WxWorkCallback 企业微信授权回调
// @Summary      企业微信授权回调
// @Description  企业微信 OAuth2 授权回调，换取用户身份信息并跳转到登录页
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        code   query  string  true   "企微授权 code"
// @Param        state  query  string  false  "自定义 state 参数"
// @Success      302  "重定向到登录页"
// @Failure      400  {object}  APIResponse  "缺少 code 参数"
// @Failure      500  {object}  APIResponse  "认证失败"
// @Router       /api/auth/wxwork/callback [get]
func (h *AuthHandler) WxWorkCallback(c *gin.Context) {
	code := c.Query("code")
	_ = c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少 code 参数"})
		return
	}

	ident, err := h.svc.Auth.WxWorkAuth(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "企业微信认证失败: " + err.Error()})
		return
	}

	token := service.GenerateToken()
	service.SaveToken(token, *ident)

	ttl := int(service.TokenTTL().Seconds())
	c.SetCookie("exam_token", token, ttl, "/", "", false, true)

	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	redirectURL := scheme + "://" + c.Request.Host + "/?token=" + token + "&role=student"
	c.Redirect(http.StatusFound, redirectURL)
}

// LoginByPassword 账号密码登录
// @Summary      账号密码登录
// @Description  使用账号+密码登录，支持管理员和员工身份
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body  req.LoginReq  true  "登录请求"
// @Success      200   {object}  APIResponse{data=LoginResponseData}  "登录成功"
// @Failure      400   {object}  APIResponse  "参数错误"
// @Failure      401   {object}  APIResponse  "账号或密码错误"
// @Router       /api/auth/login [post]
func (h *AuthHandler) LoginByPassword(c *gin.Context) {
	var r req.LoginReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var ident *service.LoginIdentity
	var err error

	if r.Role == "admin" {
		ident, err = h.svc.Auth.AdminLogin(r.WorkNo, r.Password)
	} else {
		ident, err = h.svc.Auth.StudentLogin(r.WorkNo, r.Password)
	}

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	token := service.GenerateToken()
	service.SaveToken(token, *ident)

	ttl := int(service.TokenTTL().Seconds())
	c.SetCookie("exam_token", token, ttl, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{
		"user": gin.H{
			"id":       ident.ID,
			"workNo":   ident.WorkNo,
			"name":     ident.Name,
			"wxUserId": ident.WxUserID,
			"avatar":   ident.Avatar,
			"source":   ident.Source,
			"isAdmin":  ident.IsAdmin,
		},
		"token":    token,
		"expireAt": time.Now().Add(service.TokenTTL()).Unix(),
	}, "selectedRole": r.Role, "message": "登录成功"})
}

// GetCurrentUser 获取当前用户信息
// @Summary      获取当前登录用户
// @Description  根据 Token 返回当前登录的管理员或员工信息
// @Tags         认证
// @Accept       json
// @Produce      json
// @Success      200  {object}  APIResponse{data=LoginUserDTO}  "当前用户信息"
// @Failure      401  {object}  APIResponse  "未登录或用户不存在"
// @Security     BearerToken
// @Router       /api/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	ident := GetCurrentIdentity(c)
	if ident == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	if ident.IsAdmin {
		user, err := h.svc.Auth.GetAdminInfo(ident.ID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{
			"id":      user.ID,
			"workNo":  user.WorkNo,
			"name":    user.Name,
			"isAdmin": true,
		}})
		return
	}

	student, err := h.svc.Auth.GetStudentInfo(ident.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "员工不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{
		"id":       student.ID,
		"workNo":   student.WorkNo,
		"name":     student.Name,
		"wxUserId": student.WxUserID,
		"avatar":   student.Avatar,
		"source":   student.Source,
		"isAdmin":  false,
	}})
}

// ChangePassword 修改密码
// @Summary      修改当前用户密码
// @Description  验证旧密码后修改为新密码
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body  req.ChangePasswordReq  true  "新旧密码"
// @Success      200   {object}  APIResponse  "密码修改成功"
// @Failure      400   {object}  APIResponse  "参数错误或验证失败"
// @Failure      401   {object}  APIResponse  "未登录"
// @Security     BearerToken
// @Router       /api/auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	ident := GetCurrentIdentity(c)
	if ident == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	var r req.ChangePasswordReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if r.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "新密码不能为空"})
		return
	}

	var err error
	if ident.IsAdmin {
		err = h.svc.Auth.ChangeAdminPassword(ident.ID, r.OldPassword, r.NewPassword)
	} else {
		err = h.svc.Auth.ChangeStudentPassword(ident.ID, r.OldPassword, r.NewPassword)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "密码修改成功"})
}

// Logout 退出登录
// @Summary      退出登录
// @Description  清除服务端 Token 和浏览器 Cookie
// @Tags         认证
// @Accept       json
// @Produce      json
// @Success      200  {object}  APIResponse  "已登出"
// @Security     BearerToken
// @Router       /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token, _ := c.Get("currentToken")
	if token != nil {
		service.DeleteToken(token.(string))
	}
	c.SetCookie("exam_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已登出"})
}
