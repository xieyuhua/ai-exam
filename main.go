// @title           海斛集团 API
// @version         2.0
// @description     海斛集团内部员工平台，支持考试管理、员工管理、成绩统计与导出。
// @description     新增投票系统、调查问卷系统，考题选项支持图片和视频。
// @description     认证方式：Bearer Token（通过 /api/auth/login 获取，Cookie 或 Authorization Header 传递）
// @termsOfService  http://svip.com/terms

// @contact.name   API 支持
// @contact.email  admin@svip.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey  BearerToken
// @in                          header
// @name                        Authorization
// @description                 输入 "Bearer {token}" 或通过 Cookie exam_token 传递

// @tag.name  认证
// @tag.description  登录、登出、企业微信 OAuth、修改密码

// @tag.name  分类管理
// @tag.description  考试分类的 CRUD（管理员）

// @tag.name  题目管理
// @tag.description  考题 CRUD、批量导入、模板下载（管理员）

// @tag.name  考试管理
// @tag.description  考试 CRUD、考试内考题配置（管理员）

// @tag.name  成绩管理
// @tag.description  成绩查询与导出（管理员）

// @tag.name  员工管理
// @tag.description  员工 CRUD、批量导入（管理员）

// @tag.name  投票管理
// @tag.description  投票 CRUD、选项管理（管理员）

// @tag.name  问卷管理
// @tag.description  问卷 CRUD、题目与选项管理、统计（管理员）

// @tag.name  员工端
// @tag.description  员工参加考试、查看成绩、参与投票、填写问卷

package main

import (
	"exam-system/cache"
	"exam-system/config"
	"exam-system/database"
	"exam-system/handler"
	"exam-system/repository"
	"exam-system/service"
	"log"

	_ "exam-system/docs" // swagger docs（由 swag init 生成）

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// ===== 加载配置 =====
	config.Load("config.yaml")

	// ===== 初始化数据库 =====
	database.Init(database.Config{
		Driver: config.Cfg.Database.Driver,
		DSN:    config.Cfg.Database.DSN,
	})

	// ===== 初始化 Redis 缓存 =====
	cache.Init(cache.Config{
		Host:     config.Cfg.Redis.Host,
		Port:     config.Cfg.Redis.Port,
		Password: config.Cfg.Redis.Password,
		DB:       config.Cfg.Redis.DB,
	})
	defer cache.Close()

	// ===== 初始化 DB =====
	db := database.GetDB()

	// ===== 初始化分层架构 =====
	// Repository 层（数据访问）
	catRepo := repository.NewCategoryRepo(db)
	qRepo := repository.NewQuestionRepo(db)
	examRepo := repository.NewExamRepo(db)
	eqRepo := repository.NewExamQuestionRepo(db)
	scoreRepo := repository.NewScoreRepo(db)
	studentRepo := repository.NewStudentRepo(db)
	userRepo := repository.NewUserRepo(db)
	voteRepo := repository.NewVoteRepo(db)
	surveyRepo := repository.NewSurveyRepo(db)

	// Service 层（业务逻辑）
	services := service.InitServices(catRepo, qRepo, examRepo, eqRepo, scoreRepo, studentRepo, userRepo, voteRepo, surveyRepo)

	// Handler 层（HTTP 处理 —— 结构体方式）
	h := handler.InitHandlers(services)

	// ===== Gin 引擎 =====
	r := gin.Default()

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// ===== Swagger UI =====
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 静态文件 - 旧版前端（保留向后兼容）
	r.Static("/static", "./static")
	r.StaticFile("/old.html", "./static/login.html")

	// 静态文件 - Vue SPA 前端
	r.Static("/assets", "./static-vue/assets")
	r.StaticFile("/favicon.ico", "./static-vue/favicon.ico")

	// SPA fallback — Vue Router history 模式
	r.NoRoute(func(c *gin.Context) {
		// API 路径不处理
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"code": 404, "message": "接口不存在"})
			return
		}
		c.File("./static-vue/index.html")
	})

	// ===== 认证 API（无需登录） =====
	auth := r.Group("/api/auth")
	{
		auth.GET("/wxwork/url", h.Auth.GetWxWorkAuthURL)
		auth.GET("/wxwork/callback", h.Auth.WxWorkCallback)
		auth.POST("/login", h.Auth.LoginByPassword)
	}

	// ===== 需要登录的 API =====
	authRequired := r.Group("/api")
	authRequired.Use(handler.AuthRequired)
	{
		authRequired.GET("/auth/me", h.Auth.GetCurrentUser)
		authRequired.POST("/auth/logout", h.Auth.Logout)
		authRequired.PUT("/auth/password", h.Auth.ChangePassword)
	}

	// ===== 管理端 API（需登录 + 管理员权限） =====
	admin := r.Group("/api/admin")
	admin.Use(handler.AuthRequired, handler.AdminRequired)
	{
		// 分类管理
		admin.GET("/categories", h.Category.GetCategories)
		admin.POST("/categories", h.Category.CreateCategory)
		admin.PUT("/categories/:id", h.Category.UpdateCategory)
		admin.DELETE("/categories/:id", h.Category.DeleteCategory)

		// 题目管理
		admin.GET("/questions", h.Question.GetQuestions)
		admin.POST("/questions", h.Question.CreateQuestion)
		admin.PUT("/questions/:id", h.Question.UpdateQuestion)
		admin.DELETE("/questions/:id", h.Question.DeleteQuestion)
		admin.POST("/questions/import", h.Question.ImportQuestions)
		admin.GET("/template", h.Question.DownloadTemplate)

		// 考试管理
		admin.GET("/exams", h.Exam.GetExams)
		admin.POST("/exams", h.Exam.CreateExam)
		admin.PUT("/exams/:id", h.Exam.UpdateExam)
		admin.DELETE("/exams/:id", h.Exam.DeleteExam)

		// 考试内考题管理
		admin.GET("/exams/:id/questions", h.Exam.GetExamQuestionsDetail)
		admin.POST("/exams/:id/questions", h.Exam.AddExamQuestions)
		admin.POST("/exams/:id/questions/import", h.Exam.ImportExamQuestions)
		admin.GET("/exams/:id/questions/export", h.Exam.ExportExamQuestions)
		admin.PUT("/exams/:id/questions/:qid", h.Exam.UpdateExamQuestionScore)
		admin.DELETE("/exams/:id/questions/:qid", h.Exam.RemoveExamQuestion)

		// 可导入题库题目
		admin.GET("/questions/available", h.Exam.GetAvailableQuestions)

		// 成绩管理
		admin.GET("/scores", h.Score.GetScores)
		admin.GET("/exams/:id/scores/export", h.Score.ExportExamScores)

		// 员工管理
		admin.GET("/students", h.Student.GetStudents)
		admin.POST("/students", h.Student.CreateStudent)
		admin.PUT("/students/:id", h.Student.UpdateStudent)
		admin.POST("/students/batch", h.Student.BatchCreateStudents)
		admin.POST("/students/import", h.Student.ImportStudentsExcel)
		admin.DELETE("/students/:id", h.Student.DeleteStudent)

		// 投票管理
		admin.GET("/votes", h.Vote.GetVotes)
		admin.POST("/votes", h.Vote.CreateVote)
		admin.GET("/votes/export", h.Vote.ExportVotes)
		admin.GET("/votes/:id/export", h.Vote.ExportVoteDetail)
		admin.GET("/votes/:id", h.Vote.GetVoteDetail)
		admin.PUT("/votes/:id", h.Vote.UpdateVote)
		admin.DELETE("/votes/:id", h.Vote.DeleteVote)

		// 问卷管理
		admin.GET("/surveys", h.Survey.GetSurveys)
		admin.POST("/surveys", h.Survey.CreateSurvey)
		admin.GET("/surveys/export", h.Survey.ExportSurveys)
		admin.GET("/surveys/:id/export", h.Survey.ExportSurveyDetail)
		admin.GET("/surveys/:id", h.Survey.GetSurveyDetail)
		admin.PUT("/surveys/:id", h.Survey.UpdateSurvey)
		admin.DELETE("/surveys/:id", h.Survey.DeleteSurvey)
		admin.GET("/surveys/:id/statistics", h.Survey.GetSurveyStatistics)
	}

	// ===== 员工端 API（需登录 + 非管理员） =====
	student := r.Group("/api/student")
	student.Use(handler.AuthRequired, handler.StudentOnly)
	{
		student.GET("/exams", h.Score.GetStudentExams)
		student.GET("/records", h.Score.GetStudentRecords)

		// 投票
		student.GET("/votes", h.Vote.GetVotesForStudent)
		student.GET("/votes/:id", h.Vote.GetVoteDetailForStudent)
		student.POST("/votes/submit", h.Vote.SubmitVote)

		// 问卷
		student.GET("/surveys", h.Survey.GetSurveysForStudent)
		student.GET("/surveys/:id", h.Survey.GetSurveyDetailForStudent)
		student.POST("/surveys/submit", h.Survey.SubmitSurvey)
	}

	// 考试相关 API（需登录 + 非管理员）
	exam := r.Group("/api/exam")
	exam.Use(handler.AuthRequired, handler.StudentOnly)
	{
		exam.GET("/questions", h.Score.GetExamQuestions)
		exam.POST("/submit", h.Score.SubmitExam)
	}

	// 启动服务
	port := config.Cfg.Server.Port
	log.Printf("🚀 海斛集团系统已启动: http://0.0.0.0:%s", port)
	log.Printf("📖 Swagger 文档: http://0.0.0.0:%s/swagger/index.html", port)
	log.Println("📦 架构: Repository → Service → Handler (三层分离 + 面向对象)")
	log.Println("🗳️  新增功能: 投票系统 + 调查问卷 + 富媒体选项")
	r.Run(":" + port)
}
