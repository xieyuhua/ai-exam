package repository

import (
	"exam-system/models"

	"gorm.io/gorm"
)

type QuestionRepo struct {
	db *gorm.DB
}

func NewQuestionRepo(db *gorm.DB) *QuestionRepo {
	return &QuestionRepo{db: db}
}

func (r *QuestionRepo) List(categoryID, keyword, qType string, page, pageSize int) ([]models.Question, int64, error) {
	query := r.db.Model(&models.Question{})
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	if qType != "" {
		query = query.Where("type = ?", qType)
	}

	var total int64
	query.Count(&total)

	var list []models.Question
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *QuestionRepo) GetByID(id uint) (*models.Question, error) {
	var q models.Question
	err := r.db.First(&q, id).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *QuestionRepo) GetByIDs(ids []uint) ([]models.Question, error) {
	var list []models.Question
	err := r.db.Where("id IN ?", ids).Find(&list).Error
	return list, err
}

func (r *QuestionRepo) Create(q *models.Question) error {
	return r.db.Create(q).Error
}

func (r *QuestionRepo) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Question{}).Where("id = ?", id).Updates(updates).Error
}

func (r *QuestionRepo) Delete(id uint) error {
	return r.db.Delete(&models.Question{}, id).Error
}

func (r *QuestionRepo) DeleteByCategoryID(categoryID uint) error {
	return r.db.Where("category_id = ?", categoryID).Delete(&models.Question{}).Error
}

func (r *QuestionRepo) CountByIDs(ids []uint) int64 {
	var count int64
	r.db.Model(&models.Question{}).Where("id IN ?", ids).Count(&count)
	return count
}

func (r *QuestionRepo) AvailableForExam(categoryID uint, existingIDs []uint, keyword, qType string, limit int) ([]models.Question, error) {
	query := r.db.Model(&models.Question{}).Where("category_id = ?", categoryID)
	if len(existingIDs) > 0 {
		query = query.Where("id NOT IN ?", existingIDs)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	if qType != "" {
		query = query.Where("type = ?", qType)
	}

	var list []models.Question
	err := query.Order("id desc").Limit(limit).Find(&list).Error
	return list, err
}
