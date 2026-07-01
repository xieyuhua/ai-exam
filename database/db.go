package database

import (
	"exam-system/models"
	"exam-system/util"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Config struct {
	Driver string // "sqlite" or "mysql"
	DSN    string // sqlite: file path; mysql: user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True
}

func Init(cfg Config) {
	var err error
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	default:
		dsn := cfg.DSN
		if dsn == "" {
			dsn = "exam.db"
		}
		dialector = sqlite.Open(dsn)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 自动迁移
	err = DB.AutoMigrate(
		&models.User{},
		&models.Student{},
		&models.Category{},
		&models.Question{},
		&models.Exam{},
		&models.ExamQuestion{},
		&models.Score{},
		&models.Vote{},
		&models.VoteOption{},
		&models.VoteRecord{},
		&models.Survey{},
		&models.SurveyQuestion{},
		&models.SurveyOption{},
		&models.SurveyAnswer{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	log.Println("数据库初始化成功")

	// 如果使用 SQLite，检查是否需要初始化示例数据
	if cfg.Driver == "" || cfg.Driver == "sqlite" {
		seedIfEmpty()
	}
}

func GetDB() *gorm.DB {
	return DB
}

// seedIfEmpty 首次启动时插入示例数据
func seedIfEmpty() {
	var count int64
	DB.Model(&models.Category{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("检测到空数据库，正在插入示例数据...")

	// 分类
	cats := []models.Category{
		{Name: "前端开发", Desc: "HTML、CSS、JavaScript 相关"},
		{Name: "Java 基础", Desc: "Java 语言基础与面向对象"},
		{Name: "数据库原理", Desc: "SQL、数据库设计基础"},
	}
	for i := range cats {
		DB.Create(&cats[i])
	}

	// 考题
	questions := []models.Question{
		{CategoryID: cats[0].ID, Type: "single", Title: "以下哪个不是 JavaScript 的基本数据类型？", Options: `{"a":"String","b":"Number","c":"Array","d":"Boolean"}`, Answer: `["c"]`, Explanation: "Array 是引用类型，不是基本数据类型。JS 的 7 种基本类型：String、Number、Boolean、Null、Undefined、Symbol、BigInt。"},
		{CategoryID: cats[0].ID, Type: "multiple", Title: "以下哪些是 CSS 的布局方式？（多选）", Options: `{"a":"Flexbox","b":"Grid","c":"Float","d":"Position"}`, Answer: `["a","b","c","d"]`, Explanation: "四种都是 CSS 布局方式。Flexbox 弹性布局，Grid 网格布局，Float 浮动布局，Position 定位布局。"},
		{CategoryID: cats[0].ID, Type: "single", Title: "HTML5 中用于绘制图形的标签是？", Options: `{"a":"<svg>","b":"<canvas>","c":"<graphic>","d":"<draw>"}`, Answer: `["b"]`, Explanation: "canvas 标签用于通过 JavaScript 绘制图形，svg 用于矢量图形。"},
		{CategoryID: cats[1].ID, Type: "single", Title: "Java 中哪个关键字用于类继承？", Options: `{"a":"implement","b":"extends","c":"inherit","d":"using"}`, Answer: `["b"]`, Explanation: "extends 用于类继承，implements 用于接口实现。"},
		{CategoryID: cats[1].ID, Type: "multiple", Title: "以下哪些是 Java 的基本数据类型？（多选）", Options: `{"a":"int","b":"String","c":"boolean","d":"double"}`, Answer: `["a","c","d"]`, Explanation: "Java 8种基本类型：byte、short、int、long、float、double、char、boolean。String 是引用类型。"},
		{CategoryID: cats[2].ID, Type: "single", Title: "SQL 中用于查询数据的关键字是？", Options: `{"a":"GET","b":"FIND","c":"SELECT","d":"QUERY"}`, Answer: `["c"]`, Explanation: "SELECT 是 SQL 的数据查询语言关键字。"},
	}
	for i := range questions {
		DB.Create(&questions[i])
	}

	// 考试
	now := time.Now()
	exam1 := models.Exam{
		CategoryID:    cats[0].ID,
		Title:         "前端开发能力测试",
		Duration:      30,
		TotalScore:    100,
		StartTime:     now.Add(-24 * time.Hour),
		EndTime:       now.Add(24 * 7 * time.Hour),
		CanViewAnswer: boolPtr(true),
		Status:        "active",
	}
	exam2 := models.Exam{
		CategoryID:    cats[0].ID,
		Title:         "CSS 专项练习",
		Duration:      15,
		TotalScore:    100,
		StartTime:     now.Add(-72 * time.Hour),
		EndTime:       now.Add(-24 * time.Hour),
		CanViewAnswer: boolPtr(true),
		Status:        "ended",
	}
	exam3 := models.Exam{
		CategoryID:    cats[1].ID,
		Title:         "Java 基础测试",
		Duration:      45,
		TotalScore:    100,
		StartTime:     now.Add(48 * time.Hour),
		EndTime:       now.Add(24 * 7 * time.Hour),
		CanViewAnswer: boolPtr(false),
		Status:        "upcoming",
	}
	DB.Create(&exam1)
	DB.Create(&exam2)
	DB.Create(&exam3)

	// 关联考题
	DB.Create(&models.ExamQuestion{ExamID: exam1.ID, QuestionID: questions[0].ID})
	DB.Create(&models.ExamQuestion{ExamID: exam1.ID, QuestionID: questions[1].ID})
	DB.Create(&models.ExamQuestion{ExamID: exam1.ID, QuestionID: questions[2].ID})
	DB.Create(&models.ExamQuestion{ExamID: exam2.ID, QuestionID: questions[1].ID})
	DB.Create(&models.ExamQuestion{ExamID: exam2.ID, QuestionID: questions[2].ID})
	DB.Create(&models.ExamQuestion{ExamID: exam3.ID, QuestionID: questions[3].ID})
	DB.Create(&models.ExamQuestion{ExamID: exam3.ID, QuestionID: questions[4].ID})

	// 默认管理员账号
	adminUser := models.User{
		WorkNo:   "admin",
		Name:     "系统管理员",
		Password: util.HashPassword("admin123"),
	}
	DB.Create(&adminUser)
	log.Println("默认管理员已创建: 账号=admin, 密码=admin123")

	// 示例员工
	exampleStudents := []models.Student{
		{WorkNo: "stu001", Name: "张三", Password: util.HashPassword("123456"), Source: "import"},
		{WorkNo: "stu002", Name: "李四", Password: util.HashPassword("123456"), Source: "import"},
	}
	for i := range exampleStudents {
		DB.Create(&exampleStudents[i])
	}
	log.Println("示例员工已创建: 账号=stu001/stu002, 密码=123456")

	log.Println("示例数据初始化完成")
}

func boolPtr(b bool) *bool {
	return &b
}
