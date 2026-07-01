package repository

import (
	"exam-system/models"
	"time"

	"gorm.io/gorm"
)

type SurveyRepo struct {
	db *gorm.DB
}

func NewSurveyRepo(db *gorm.DB) *SurveyRepo {
	return &SurveyRepo{db: db}
}

func (r *SurveyRepo) List(keyword, status string, page, pageSize int) ([]models.Survey, int64, error) {
	query := r.db.Model(&models.Survey{})
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	var list []models.Survey
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *SurveyRepo) GetByID(id uint) (*models.Survey, error) {
	var s models.Survey
	err := r.db.First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SurveyRepo) Create(s *models.Survey) error {
	return r.db.Create(s).Error
}

func (r *SurveyRepo) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Survey{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SurveyRepo) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var qIDs []uint
		tx.Model(&models.SurveyQuestion{}).Where("survey_id = ?", id).Pluck("id", &qIDs)
		if len(qIDs) > 0 {
			if err := tx.Where("survey_question_id IN ?", qIDs).Delete(&models.SurveyOption{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("survey_id = ?", id).Delete(&models.SurveyQuestion{}).Error; err != nil {
			return err
		}
		if err := tx.Where("survey_id = ?", id).Delete(&models.SurveyAnswer{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Survey{}, id).Error
	})
}

func (r *SurveyRepo) UpdateStatus(ids []uint, status string) error {
	return r.db.Model(&models.Survey{}).Where("id IN ?", ids).Update("status", status).Error
}

func (r *SurveyRepo) AutoUpdateStatus() {
	// SQLite 不存时区，传格式化字符串（无时区后缀）确保字符串比较一致
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	r.db.Model(&models.Survey{}).Where("status = ? AND start_time <= ? AND end_time >= ?", "upcoming", nowStr, nowStr).Update("status", "active")
	r.db.Model(&models.Survey{}).Where("status = ? AND end_time < ?", "active", nowStr).Update("status", "ended")
}

// SurveyQuestion
func (r *SurveyRepo) CreateQuestions(questions []models.SurveyQuestion) error {
	return r.db.Create(&questions).Error
}

func (r *SurveyRepo) GetQuestions(surveyID uint) ([]models.SurveyQuestion, error) {
	var questions []models.SurveyQuestion
	err := r.db.Where("survey_id = ?", surveyID).Order("sort_order asc, id asc").Find(&questions).Error
	return questions, err
}

func (r *SurveyRepo) DeleteQuestions(surveyID uint) error {
	return r.db.Where("survey_id = ?", surveyID).Delete(&models.SurveyQuestion{}).Error
}

// SurveyOption
func (r *SurveyRepo) CreateOptions(options []models.SurveyOption) error {
	return r.db.Create(&options).Error
}

func (r *SurveyRepo) GetOptions(questionIDs []uint) ([]models.SurveyOption, error) {
	var opts []models.SurveyOption
	err := r.db.Where("survey_question_id IN ?", questionIDs).Order("sort_order asc, id asc").Find(&opts).Error
	return opts, err
}

func (r *SurveyRepo) DeleteOptionsByQIDs(questionIDs []uint) error {
	return r.db.Where("survey_question_id IN ?", questionIDs).Delete(&models.SurveyOption{}).Error
}

// SurveyAnswer
func (r *SurveyRepo) CreateAnswers(answers []models.SurveyAnswer) error {
	return r.db.Create(&answers).Error
}

func (r *SurveyRepo) GetAnswer(surveyID uint, studentName string) ([]models.SurveyAnswer, error) {
	var answers []models.SurveyAnswer
	err := r.db.Where("survey_id = ? AND student_name = ?", surveyID, studentName).Find(&answers).Error
	return answers, err
}

func (r *SurveyRepo) CountCompleted(surveyID uint) (int64, error) {
	var count int64
	// 统计至少回答了一题的不同员工数
	err := r.db.Model(&models.SurveyAnswer{}).Where("survey_id = ?", surveyID).Distinct("student_name").Count(&count).Error
	return count, err
}

func (r *SurveyRepo) CountTotal() (int64, error) {
	var count int64
	err := r.db.Model(&models.Survey{}).Count(&count).Error
	return count, err
}

func (r *SurveyRepo) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&models.Survey{}).Where("status = ?", "active").Count(&count).Error
	return count, err
}

func (r *SurveyRepo) GetQuestionByID(id uint) (*models.SurveyQuestion, error) {
	var q models.SurveyQuestion
	err := r.db.First(&q, id).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// GetAllAnswers 获取某个问卷的全部答案（不过滤学员）
func (r *SurveyRepo) GetAllAnswers(surveyID uint) ([]models.SurveyAnswer, error) {
	var answers []models.SurveyAnswer
	err := r.db.Where("survey_id = ?", surveyID).Find(&answers).Error
	return answers, err
}
