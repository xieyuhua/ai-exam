package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"exam-system/service"
)

// AuthRequired 认证中间件 —— 从缓存读取身份
func AuthRequired(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录"})
		c.Abort()
		return
	}

	ident, ok := service.GetTokenIdentity(token)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "登录已过期，请重新登录"})
		c.Abort()
		return
	}

	c.Set("currentIdentity", ident)
	c.Set("currentToken", token)

	// 刷新 token 有效期
	service.RefreshToken(token, *ident)

	c.Next()
}

// AdminRequired 管理员权限中间件
func AdminRequired(c *gin.Context) {
	ident, exists := c.Get("currentIdentity")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录"})
		c.Abort()
		return
	}
	id := ident.(*service.LoginIdentity)
	if !id.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无管理员权限"})
		c.Abort()
		return
	}
	c.Next()
}

// StudentOnly 员工专属中间件
func StudentOnly(c *gin.Context) {
	ident, exists := c.Get("currentIdentity")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录"})
		c.Abort()
		return
	}
	id := ident.(*service.LoginIdentity)
	if id.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "管理员账号不能参与员工端功能"})
		c.Abort()
		return
	}
	c.Next()
}

// GetCurrentIdentity 从上下文中获取当前登录身份
func GetCurrentIdentity(c *gin.Context) *service.LoginIdentity {
	if ident, exists := c.Get("currentIdentity"); exists {
		return ident.(*service.LoginIdentity)
	}
	return nil
}

// extractToken 提取 token
func extractToken(c *gin.Context) string {
	// Authorization header
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	// Cookie
	if token, err := c.Cookie("exam_token"); err == nil {
		return token
	}
	// Query
	return c.Query("token")
}
