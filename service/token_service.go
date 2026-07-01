package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"exam-system/cache"
	"exam-system/config"
)

// ==================== Token 管理 ====================

func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func SaveToken(token string, identity LoginIdentity) {
	cache.Set("token:"+token, identity, tokenTTL)
}

func GetTokenIdentity(token string) (*LoginIdentity, bool) {
	var ident LoginIdentity
	if !cache.GetJSON("token:"+token, &ident) {
		return nil, false
	}
	return &ident, true
}

func DeleteToken(token string) {
	cache.Del("token:" + token)
}

func RefreshToken(token string, identity LoginIdentity) {
	cache.Set("token:"+token, identity, tokenTTL)
}

// ==================== 企业微信 API ====================

type wxAccessTokenResp struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type wxUserIDResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	UserID  string `json:"UserId"`
	OpenID  string `json:"OpenId"`
}

type wxUserInfoResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	UserID  string `json:"userid"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Mobile  string `json:"mobile"`
	Email   string `json:"email"`
}

func getWxWorkAccessToken() (string, error) {
	if val, ok := cache.GetString("wx_access_token"); ok {
		return val, nil
	}

	corpID := config.Cfg.WxWork.CorpID
	secret := config.Cfg.WxWork.Secret

	resp, err := http.Get(fmt.Sprintf(
		"https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		corpID, secret,
	))
	if err != nil {
		return "", fmt.Errorf("请求 access_token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenResp wxAccessTokenResp
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析 access_token 失败: %w", err)
	}
	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("获取 access_token 失败: %s (errcode=%d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	cache.Set("wx_access_token", tokenResp.AccessToken, time.Duration(tokenResp.ExpiresIn-300)*time.Second)
	return tokenResp.AccessToken, nil
}

func getWxWorkUserID(code string) (string, error) {
	accessToken, err := getWxWorkAccessToken()
	if err != nil {
		return "", err
	}

	resp, err := http.Get(fmt.Sprintf(
		"https://qyapi.weixin.qq.com/cgi-bin/auth/getuserinfo?access_token=%s&code=%s",
		accessToken, code,
	))
	if err != nil {
		return "", fmt.Errorf("请求 getuserinfo 失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var userIDResp wxUserIDResp
	if err := json.Unmarshal(body, &userIDResp); err != nil {
		return "", fmt.Errorf("解析 getuserinfo 失败: %w", err)
	}
	if userIDResp.ErrCode != 0 {
		return "", fmt.Errorf("获取 UserID 失败: %s (errcode=%d)", userIDResp.ErrMsg, userIDResp.ErrCode)
	}
	return userIDResp.UserID, nil
}

func getWxWorkUserInfo(userID string) (*wxUserInfoResp, error) {
	accessToken, err := getWxWorkAccessToken()
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(fmt.Sprintf(
		"https://qyapi.weixin.qq.com/cgi-bin/user/get?access_token=%s&userid=%s",
		accessToken, userID,
	))
	if err != nil {
		return nil, fmt.Errorf("请求用户信息失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var userInfo wxUserInfoResp
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}
	if userInfo.ErrCode != 0 {
		return nil, fmt.Errorf("获取用户信息失败: %s (errcode=%d)", userInfo.ErrMsg, userInfo.ErrCode)
	}
	return &userInfo, nil
}

// TokenTTL 返回 token 有效期
func TokenTTL() time.Duration {
	return tokenTTL
}
