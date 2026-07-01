package service

import (
	"exam-system/repository"
	"exam-system/util"
	"fmt"
	"time"

	"exam-system/cache"
)

// ==================== 缓存适配 ====================

type CacheStore interface {
	Set(key string, value interface{}, ttl time.Duration)
	Get(key string) (interface{}, bool)
	Del(key string)
	Exists(key string) bool
	GetString(key string) (string, bool)
	GetJSON(key string, dest interface{}) bool
}

type cacheAdapter struct{}

func (c *cacheAdapter) Set(key string, value interface{}, ttl time.Duration) {
	cache.Set(key, value, ttl)
}

func (c *cacheAdapter) Get(key string) (interface{}, bool) {
	return cache.Get(key)
}

func (c *cacheAdapter) Del(key string) {
	cache.Del(key)
}

func (c *cacheAdapter) Exists(key string) bool {
	return cache.Exists(key)
}

func (c *cacheAdapter) GetString(key string) (string, bool) {
	return cache.GetString(key)
}

func (c *cacheAdapter) GetJSON(key string, dest interface{}) bool {
	return cache.GetJSON(key, dest)
}

// ==================== 工具函数 ====================

func HashPassword(password string) string {
	return util.HashPassword(password)
}

func parseTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unknown time format: %s", s)
}

// ==================== 导入类型 ====================

type ImportQuestionRow struct {
	Type        string
	Title       string
	Options     map[string]string
	Answer      []string
	Explanation string
	Score       float64 // 导入分值（0 表示使用默认值）
}

// ==================== ServiceRegistry 聚合所有 service ====================

type ServiceRegistry struct {
	Category *CategoryService
	Question *QuestionService
	Exam     *ExamService
	Score    *ScoreService
	Student  *StudentService
	Auth     *AuthService
	Vote     *VoteService
	Survey   *SurveyService
}

// InitServices 初始化所有服务
func InitServices(
	catRepo *repository.CategoryRepo,
	qRepo *repository.QuestionRepo,
	examRepo *repository.ExamRepo,
	eqRepo *repository.ExamQuestionRepo,
	scoreRepo *repository.ScoreRepo,
	studentRepo *repository.StudentRepo,
	userRepo *repository.UserRepo,
	voteRepo *repository.VoteRepo,
	surveyRepo *repository.SurveyRepo,
) *ServiceRegistry {
	return &ServiceRegistry{
		Category: NewCategoryService(catRepo),
		Question: NewQuestionService(qRepo, catRepo),
		Exam:     NewExamService(examRepo, eqRepo, scoreRepo, catRepo, qRepo),
		Score:    NewScoreService(scoreRepo, examRepo, catRepo),
		Student:  NewStudentService(studentRepo, userRepo),
		Auth:     NewAuthService(userRepo, studentRepo),
		Vote:     NewVoteService(voteRepo),
		Survey:   NewSurveyService(surveyRepo),
	}
}
