package handler

import (
	"exam-system/database"
	"exam-system/service"

	"gorm.io/gorm"
)

// ==================== HandlerRegistry 聚合所有 handler ====================

type HandlerRegistry struct {
	Auth     *AuthHandler
	Category *CategoryHandler
	Question *QuestionHandler
	Exam     *ExamHandler
	Score    *ScoreHandler
	Student  *StudentHandler
	Vote     *VoteHandler
	Survey   *SurveyHandler
}

// InitHandlers 创建所有 handler，注入 service 和 db
func InitHandlers(sr *service.ServiceRegistry) *HandlerRegistry {
	db := database.GetDB()
	return &HandlerRegistry{
		Auth:     NewAuthHandler(sr),
		Category: NewCategoryHandler(sr),
		Question: NewQuestionHandler(sr, db),
		Exam:     NewExamHandler(sr, db),
		Score:    NewScoreHandler(sr, db),
		Student:  NewStudentHandler(sr),
		Vote:     NewVoteHandler(sr, db),
		Survey:   NewSurveyHandler(sr, db),
	}
}

// ==================== 基础抽象 ====================

// baseHandler 提供通用工具
type baseHandler struct {
	svc *service.ServiceRegistry
	db  *gorm.DB
}

// getDB 安全获取 DB 实例（db 字段可能为 nil 时降级）
func (b *baseHandler) getDB() *gorm.DB {
	if b.db != nil {
		return b.db
	}
	return database.GetDB()
}
