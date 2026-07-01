# 在线考试系统 v2.0

基于 Go 三层分层架构 + Vue 3 SPA 前端的综合在线平台，支持考试管理、投票系统、调查问卷系统。

---

## 一、技术栈

| 层级 | 技术选型 |
|------|----------|
| 后端语言 | Go 1.21 |
| Web 框架 | Gin v1.9 |
| ORM | GORM v1.25 |
| 数据库 | SQLite（开发） / MySQL（生产） |
| 缓存 | Redis + 内存降级 |
| 认证 | 工号密码 / 企业微信 OAuth 2.0 |
| 导出 | Excel（excelize/v2）/ CSV |
| 配置 | YAML |
| API 文档 | Swagger（swaggo） |
| 前端框架 | Vue 3 + Vue Router 4 |
| 构建工具 | Vite 5 |

---

## 二、系统分层架构

```
┌─────────────────────────────────────────────────────────┐
│                   Client (Browser)                       │
│             Vue 3 SPA (static-vue/)                      │
└────────────┬─────────────────────────────────────────────┘
             ▼
┌─────────────────────────────────────────────────────────┐
│                  HTTP Middleware Layer                    │
│  ┌──────────────┐ ┌──────────────┐ ┌─────────────────┐  │
│  │  CORS Filter  │ │ AuthRequired  │ │ AdminRequired / │  │
│  │  (跨域处理)    │ │ (Token 校验)  │ │ StudentOnly     │  │
│  └──────────────┘ └──────────────┘ └─────────────────┘  │
└──────────────────────┬──────────────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────────────┐
│                 Handler Layer (handler/)                  │
│  Auth / Category / Question / Exam / Score / Student /  │
│  Vote / Survey                                           │
└──────────────────────┬──────────────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────────────┐
│                 Service Layer (service/)                  │
│  Auth / Category / Question / Exam / Score / Student /  │
│  Vote / Survey / Token                                   │
└──────────────────────┬──────────────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────────────┐
│               Repository Layer (repository/)              │
│  User / Student / Category / Question / Exam /          │
│  ExamQuestion / Score / Vote / Survey                    │
└──────────────────────┬──────────────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────────────┐
│                      Data Layer                           │
│  ┌────────────┐  ┌───────────────┐  ┌───────────────┐   │
│  │   GORM     │  │  Redis Cache   │  │  DTO (dto/)  │   │
│  │ SQLite/MySQL│  │ (token/state)  │  │ request/response│  │
│  └────────────┘  └───────────────┘  └───────────────┘   │
└─────────────────────────────────────────────────────────┘
```

---

## 三、实体关系图 (ER)

```
 ┌──────────┐        ┌──────────────┐        ┌──────────────┐
 │ Category │        │   Student    │        │     User     │
 │──────────│        │──────────────│        │──────────────│
 │ id       │◄───FK──│ id           │        │ id           │
 │ name     │        │ workNo       │        │ workNo       │
 │ desc     │        │ name         │        │ name         │
 └────┬─────┘        │ password     │        │ password     │
      │              │ wxUserId     │        └──────────────┘
      │ 1:N          │ source       │
      ▼              └──────┬───────┘
 ┌──────────┐               │
 │ Question │               │
 │──────────│               │
 │ id       │               │
 │ category │──FK           │
 │ type     │               │
 │ title    │        ┌──────┴────────┐        ┌──────────────┐
 │ options  │        │    Score      │        │    Exam      │
 │ answer   │        │───────────────│        │──────────────│
 └────┬─────┘        │ id            │◄───FK──│ id           │
      │              │ studentName   │        │ category_id  │
      │ M:N          │ exam_id       │        │ title        │
      ▼              │ score         │        │ duration     │
 ┌──────────────┐    │ correct       │        │ totalScore   │
 │ ExamQuestion │    │ total         │        │ startTime    │
 │──────────────│    │ answers(JSON) │        │ endTime      │
 │ exam_id (PK) │    └───────────────┘        │ status       │
 │ question_id  │                             │ canViewAnswer│
 │ score        │                             └──────┬───────┘
 └──────────────┘                                    │
                                                     │
 ┌──────────────┐        ┌──────────────┐            │
 │    Vote      │        │   Survey     │            │
 │──────────────│        │──────────────│            │
 │ id           │        │ id           │            │
 │ title        │        │ title        │            │
 │ description  │        │ description  │            │
 │ startTime    │        │ startTime    │            │
 │ endTime      │        │ endTime      │            │
 │ voteType     │        │ status       │            │
 │ status       │        └──────┬───────┘            │
 └──────┬───────┘               │                    │
        │ 1:N                   │ 1:N                │
        ▼                       ▼                    │
 ┌──────────────┐        ┌──────────────┐            │
 │ VoteOption   │        │SurveyQuestion│            │
 │──────────────│        │──────────────│            │
 │ id           │        │ id           │            │
 │ vote_id      │        │ survey_id    │            │
 │ label        │        │ title        │            │
 │ content      │        │ type         │            │
 └──────────────┘        └──────┬───────┘            │
        │                       │ 1:N                │
        │ 1:N                   ▼                    │
        ▼                ┌──────────────┐            │
 ┌──────────────┐        │SurveyOption  │            │
 │ VoteRecord   │        │──────────────│            │
 │──────────────│        │ id           │            │
 │ id           │        │ qid          │            │
 │ vote_id      │        │ label        │            │
 │ studentName  │        │ content      │            │
 │ optionIds    │        └──────────────┘            │
 └──────────────┘        ┌──────────────┐            │
                         │SurveyAnswer  │            │
                         │──────────────│            │
                         │ id           │            │
                         │ survey_id    │            │
                         │ studentName  │            │
                         │ answer       │            │
                         └──────────────┘            │
```

---

## 四、项目文件结构

```
exam/
├── main.go                     # 入口：依赖注入 + 路由注册
├── config.yaml                 # 配置文件
├── go.mod / go.sum             # Go 模块依赖
├── README.md                   # 本文件
│
├── config/                     # 配置模块
│   └── config.go
│
├── database/                   # 数据库初始化
│   └── db.go                   # DB连接、AutoMigrate、种子数据
│
├── cache/                      # 缓存模块
│   └── cache.go                # Redis + 内存降级
│
├── models/                     # 数据模型（每个实体一个文件）
│   ├── user.go                 # User 管理员
│   ├── student.go              # Student 员工
│   ├── category.go             # Category 分类
│   ├── question.go             # Question 考题
│   ├── exam.go                 # Exam 考试
│   ├── exam_question.go        # ExamQuestion 关联表
│   ├── score.go                # Score 成绩
│   ├── vote.go                 # Vote/VoteOption/VoteRecord
│   ├── survey.go               # Survey/SurveyQuestion/SurveyOption/SurveyAnswer
│   └── models.go               # DTO/请求响应结构体
│
├── dto/                        # 数据传输对象
│   ├── request/                # 请求体定义
│   │   ├── auth.go
│   │   ├── category.go
│   │   ├── exam.go
│   │   ├── question.go
│   │   ├── score.go
│   │   ├── student.go
│   │   ├── survey.go
│   │   └── vote.go
│   └── response/               # 响应体定义
│       ├── auth.go
│       ├── category.go
│       ├── exam.go
│       ├── question.go
│       ├── score.go
│       ├── student.go
│       ├── survey.go
│       └── vote.go
│
├── repository/                 # 数据访问层
│   ├── user_repo.go
│   ├── student_repo.go
│   ├── category_repo.go
│   ├── question_repo.go
│   ├── exam_repo.go
│   ├── exam_question_repo.go
│   ├── score_repo.go
│   ├── vote_repo.go
│   └── survey_repo.go
│
├── service/                    # 业务逻辑层
│   ├── service.go              # ServiceRegistry + InitServices
│   ├── auth_service.go
│   ├── token_service.go
│   ├── category_service.go
│   ├── question_service.go
│   ├── exam_service.go
│   ├── score_service.go
│   ├── student_service.go
│   ├── vote_service.go
│   └── survey_service.go
│
├── handler/                    # HTTP 处理层
│   ├── handler.go              # Handler 聚合 + InitHandlers
│   ├── middleware.go           # AuthRequired/AdminRequired/StudentOnly
│   ├── auth_handler.go
│   ├── category_handler.go
│   ├── question_handler.go
│   ├── exam_handler.go
│   ├── score_handler.go
│   ├── student_handler.go
│   ├── vote_handler.go
│   └── survey_handler.go
│
├── util/                       # 工具函数
│   └── util.go                 # HashPassword、分页、排序等
│
├── docs/                       # Swagger 文档
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
│
├── frontend/                   # Vue 3 前端源码
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   └── src/
│       ├── App.vue
│       ├── main.js
│       ├── router/
│       ├── stores/             # 状态管理（auth）
│       ├── components/
│       │   ├── admin/          # 管理端组件
│       │   │   ├── CategoryPanel.vue
│       │   │   ├── QuestionPanel.vue
│       │   │   ├── ExamPanel.vue
│       │   │   ├── ExamQuestionManager.vue
│       │   │   ├── ScorePanel.vue
│       │   │   ├── StudentPanel.vue
│       │   │   ├── VotePanel.vue
│       │   │   └── SurveyPanel.vue
│       │   ├── common/         # 通用组件（Topbar、Modal、Toast）
│       │   └── toastState.js
│       └── views/              # 页面级组件
│           ├── LoginView.vue
│           ├── AdminView.vue
│           ├── StudentView.vue
│           ├── ExamListView.vue
│           ├── ExamView.vue
│           ├── VoteListView.vue
│           ├── VoteView.vue
│           ├── SurveyListView.vue
│           └── SurveyView.vue
│
└── static-vue/                 # 前端构建产物（Vite build 输出）
    ├── index.html
    └── assets/
```

---

## 五、API 路由一览

### 认证（无需登录）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/login` | 工号密码登录 |
| GET | `/api/auth/wxwork/url` | 获取企微授权 URL |
| GET | `/api/auth/wxwork/callback` | 企微 OAuth 回调 |

### 管理端（需登录 + 管理员）

#### 分类管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/categories` | 分类列表 |
| POST | `/api/admin/categories` | 新增分类 |
| PUT | `/api/admin/categories/:id` | 编辑分类 |
| DELETE | `/api/admin/categories/:id` | 删除分类 |

#### 题目管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/questions` | 题目列表 |
| POST | `/api/admin/questions` | 新增题目 |
| PUT | `/api/admin/questions/:id` | 编辑题目 |
| DELETE | `/api/admin/questions/:id` | 删除题目 |
| POST | `/api/admin/questions/import` | 批量导入题目 |
| GET | `/api/admin/template` | 下载导入模板 |
| GET | `/api/admin/questions/available` | 可导入题库的题目 |

#### 考试管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/exams` | 考试列表 |
| POST | `/api/admin/exams` | 新增考试 |
| PUT | `/api/admin/exams/:id` | 编辑考试 |
| DELETE | `/api/admin/exams/:id` | 删除考试 |
| GET | `/api/admin/exams/:id/questions` | 考试内题目详情 |
| POST | `/api/admin/exams/:id/questions` | 添加题目到考试 |
| POST | `/api/admin/exams/:id/questions/import` | 导入题目到考试 |
| GET | `/api/admin/exams/:id/questions/export` | 导出考题 Excel |
| PUT | `/api/admin/exams/:id/questions/:qid` | 修改题目分值 |
| DELETE | `/api/admin/exams/:id/questions/:qid` | 从考试移除题目 |

#### 成绩管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/scores` | 成绩列表 |
| GET | `/api/admin/scores/export` | 导出全部成绩 |
| GET | `/api/admin/exams/:id/scores/export` | 按考试导出成绩 |

#### 员工管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/students` | 员工列表 |
| POST | `/api/admin/students` | 新增员工 |
| PUT | `/api/admin/students/:id` | 编辑员工 |
| POST | `/api/admin/students/batch` | 批量新增 |
| POST | `/api/admin/students/import` | 批量导入 Excel |
| DELETE | `/api/admin/students/:id` | 删除员工 |

#### 投票管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/votes` | 投票列表 |
| POST | `/api/admin/votes` | 新增投票 |
| GET | `/api/admin/votes/:id` | 投票详情 |
| PUT | `/api/admin/votes/:id` | 编辑投票 |
| DELETE | `/api/admin/votes/:id` | 删除投票 |
| GET | `/api/admin/votes/:id/export` | 导出单个投票数据 |
| GET | `/api/admin/votes/export` | 导出全部投票 |

#### 问卷管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/surveys` | 问卷列表 |
| POST | `/api/admin/surveys` | 新增问卷 |
| GET | `/api/admin/surveys/:id` | 问卷详情 |
| PUT | `/api/admin/surveys/:id` | 编辑问卷 |
| DELETE | `/api/admin/surveys/:id` | 删除问卷 |
| GET | `/api/admin/surveys/:id/statistics` | 问卷统计 |
| GET | `/api/admin/surveys/:id/export` | 导出单个问卷 |
| GET | `/api/admin/surveys/export` | 导出全部问卷 |

### 员工端（需登录 + 非管理员）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/student/exams` | 可参加的考试列表 |
| GET | `/api/student/records` | 考试历史记录 |
| GET | `/api/student/votes` | 可参与的投票列表 |
| GET | `/api/student/votes/:id` | 投票详情 |
| POST | `/api/student/votes/submit` | 提交投票 |
| GET | `/api/student/surveys` | 可填写的问卷列表 |
| GET | `/api/student/surveys/:id` | 问卷详情 |
| POST | `/api/student/surveys/submit` | 提交问卷 |
| GET | `/api/exam/questions` | 获取考试题目 |
| POST | `/api/exam/submit` | 提交考试 |

---

## 六、考题类型及判分规则

| 类型 | 标识 | 选项格式 | 答案格式 | 判分规则 |
|------|------|----------|----------|----------|
| 单选题 | `single` | `{"a":"..","b":".."}` | `["a"]` | 完全匹配 |
| 多选题 | `multiple` | `{"a":"..","b":".."}` | `["a","c"]` | 排序后完全匹配 |
| 判断题 | `judge` | 自动生成对/错 | `["a"]` 或 `["b"]` | 完全匹配 |
| 填空题 | `fill` | 可为空 | `["答案"]` | Trim + EqualFold |
| 简答题 | `essay` | 可为空 | `["答案"]` | Trim + EqualFold |

---

## 七、快速开始

### 1. 编译后端

```bash
go mod tidy
go build -o exam-system.exe .
./exam-system.exe
```

### 2. 编译前端

```bash
cd frontend/
npm install
npm run build
```

前端构建产物输出到 `static-vue/` 目录，后端会自动服务。

### 3. 访问

- 前端页面：`http://localhost:8485`
- Swagger 文档：`http://localhost:8485/swagger/index.html`

### 4. 默认账号

| 角色 | 工号 | 密码 |
|------|------|------|
| 管理员 | `admin` | `admin123` |
| 考生 | `stu001` | `123456` |
| 考生 | `stu002` | `123456` |

---

## 八、配置说明

```yaml
# config.yaml
server:
  port: "8485"            # 服务端口

database:
  driver: sqlite          # sqlite | mysql
  dsn: exam.db            # SQLite: 文件路径  MySQL: user:pass@tcp(host:port)/db

redis:
  host: "127.0.0.1"       # Redis 地址
  port: "6379"
  password: ""
  db: 12

wxwork:                   # 企业微信（不配置则禁用企微登录）
  corp_id: ""
  agent_id: ""
  secret: ""
  redirect_uri: ""
```

---

## 九、配置 MySQL（生产环境）

```yaml
database:
  driver: mysql
  dsn: "root:password@tcp(127.0.0.1:3306)/exam?charset=utf8mb4&parseTime=True&loc=Local"
```

切换后 GORM AutoMigrate 会自动建表。
