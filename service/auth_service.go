package service

import (
	"exam-system/cache"
	"exam-system/config"
	"exam-system/models"
	"exam-system/repository"
	"exam-system/util"
	"net/url"
	"time"
)

type AuthService struct {
	userRepo    *repository.UserRepo
	studentRepo *repository.StudentRepo
}

func NewAuthService(userRepo *repository.UserRepo, studentRepo *repository.StudentRepo) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		studentRepo: studentRepo,
	}
}

// LoginIdentity 登录身份
type LoginIdentity struct {
	IsAdmin  bool   `json:"isAdmin"`
	ID       uint   `json:"id"`
	WorkNo   string `json:"workNo"`
	Name     string `json:"name"`
	WxUserID string `json:"wxUserId"`
	Avatar   string `json:"avatar"`
	Source   string `json:"source,omitempty"`
}

const tokenTTL = 24 * time.Hour

// AdminLogin 管理员登录
func (s *AuthService) AdminLogin(workNo, password string) (*LoginIdentity, error) {
	user, err := s.userRepo.GetByWorkNo(workNo)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !util.CheckPassword(user.Password, password) {
		return nil, ErrInvalidCredentials
	}
	return &LoginIdentity{
		IsAdmin: true,
		ID:      user.ID,
		WorkNo:  user.WorkNo,
		Name:    user.Name,
	}, nil
}

// StudentLogin 员工登录
func (s *AuthService) StudentLogin(workNo, password string) (*LoginIdentity, error) {
	student, err := s.studentRepo.GetByWorkNo(workNo)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !util.CheckPassword(student.Password, password) {
		return nil, ErrInvalidCredentials
	}
	return &LoginIdentity{
		IsAdmin:  false,
		ID:       student.ID,
		WorkNo:   student.WorkNo,
		Name:     student.Name,
		WxUserID: student.WxUserID,
		Avatar:   student.Avatar,
		Source:   student.Source,
	}, nil
}

// GetAdminInfo 获取管理员信息
func (s *AuthService) GetAdminInfo(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

// GetStudentInfo 获取员工信息
func (s *AuthService) GetStudentInfo(id uint) (*models.Student, error) {
	return s.studentRepo.GetByID(id)
}

// ChangeAdminPassword 修改管理员密码
func (s *AuthService) ChangeAdminPassword(id uint, oldPwd, newPwd string) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return ErrUserNotFound
	}
	if !util.CheckPassword(user.Password, oldPwd) {
		return ErrInvalidCredentials
	}
	return s.userRepo.UpdatePassword(id, util.HashPassword(newPwd))
}

// ChangeStudentPassword 修改员工密码
func (s *AuthService) ChangeStudentPassword(id uint, oldPwd, newPwd string) error {
	student, err := s.studentRepo.GetByID(id)
	if err != nil {
		return ErrUserNotFound
	}
	if !util.CheckPassword(student.Password, oldPwd) {
		return ErrInvalidCredentials
	}
	return s.studentRepo.Update(id, map[string]interface{}{"password": util.HashPassword(newPwd)})
}

// WxWorkAuth 企业微信OAuth认证
func (s *AuthService) WxWorkAuth(code string) (*LoginIdentity, error) {
	wxUserID, err := getWxWorkUserID(code)
	if err != nil {
		return nil, err
	}

	wxUser, err := getWxWorkUserInfo(wxUserID)
	if err != nil {
		return nil, err
	}

	student, err := s.studentRepo.GetByWxUserID(wxUserID)
	if err != nil {
		// 检查账号绑定
		student2, err2 := s.studentRepo.GetByWorkNoAndEmptyWx(wxUserID)
		if err2 == nil {
			s.studentRepo.Update(student2.ID, map[string]interface{}{
				"wx_user_id": wxUserID,
				"name":       wxUser.Name,
				"avatar":     wxUser.Avatar,
			})
			student = student2
		} else {
			// 新员工
			newStudent := models.Student{
				WorkNo:   wxUserID,
				Name:     wxUser.Name,
				WxUserID: wxUserID,
				Avatar:   wxUser.Avatar,
				Source:   "wxwork",
			}
			if err := s.studentRepo.Create(&newStudent); err != nil {
				return nil, err
			}
			student = &newStudent
		}
	} else {
		s.studentRepo.Update(student.ID, map[string]interface{}{
			"name":   wxUser.Name,
			"avatar": wxUser.Avatar,
		})
	}

	return &LoginIdentity{
		IsAdmin:  false,
		ID:       student.ID,
		WorkNo:   student.WorkNo,
		Name:     student.Name,
		WxUserID: student.WxUserID,
		Avatar:   student.Avatar,
		Source:   student.Source,
	}, nil
}

// GetWxWorkAuthURL 获取授权URL
func GetWxWorkAuthURL(host string, scheme string) (string, error) {
	c := config.Cfg.WxWork
	if c.CorpID == "" {
		return "", ErrWxWorkNotConfigured
	}

	redirectURI := c.RedirectURI
	if redirectURI == "" {
		redirectURI = scheme + "://" + host + "/api/auth/wxwork/callback"
	}

	state := GenerateToken()
	cache.Set("wx_state:"+state, "student", 5*time.Minute)

	authURL := "https://open.weixin.qq.com/connect/oauth2/authorize?appid=" + c.CorpID +
		"&redirect_uri=" + url.QueryEscape(redirectURI) +
		"&response_type=code&scope=snsapi_base&state=" + state + "#wechat_redirect"
	return authURL, nil
}

// ======================== 错误定义 ========================

var (
	ErrInvalidCredentials   = &AppError{Msg: "账号或密码错误"}
	ErrUserNotFound         = &AppError{Msg: "用户不存在"}
	ErrWxWorkNotConfigured  = &AppError{Msg: "企业微信未配置"}
)

type AppError struct {
	Msg string
}

func (e *AppError) Error() string {
	return e.Msg
}
