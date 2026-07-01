package response

// LoginUserDTO 登录用户信息
type LoginUserDTO struct {
	ID       uint   `json:"id"`
	WorkNo   string `json:"workNo"`
	Name     string `json:"name"`
	WxUserID string `json:"wxUserId,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Source   string `json:"source,omitempty"`
	IsAdmin  bool   `json:"isAdmin"`
}

// LoginResponseData 登录响应数据
type LoginResponseData struct {
	User     LoginUserDTO `json:"user"`
	Token    string       `json:"token"`
	ExpireAt int64        `json:"expireAt"`
}

// WxWorkAuthData 企业微信授权URL
type WxWorkAuthData struct {
	AuthURL string `json:"authUrl"`
}
