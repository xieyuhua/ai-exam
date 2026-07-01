package repository

import (
	"exam-system/models"

	"gorm.io/gorm"
)

type ExamRepo struct {
	db *gorm.DB
}

func NewExamRepo(db *gorm.DB) *ExamRepo {
	return &ExamRepo{db: db}
}

func (r *ExamRepo) List(categoryID, keyword, status string, page, pageSize int) ([]models.Exam, int64, error) {
	query := r.db.Model(&models.Exam{})
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var list []models.Exam
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *ExamRepo) GetByID(id uint) (*models.Exam, error) {
	var exam models.Exam
	err := r.db.First(&exam, id).Error
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *ExamRepo) CreateWithTx(tx *gorm.DB, exam *models.Exam) error {
	return tx.Create(exam).Error
}

func (r *ExamRepo) Create(exam *models.Exam) error {
	return r.db.Create(exam).Error
}

func (r *ExamRepo) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Exam{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ExamRepo) Delete(id uint) error {
	return r.db.Delete(&models.Exam{}, id).Error
}

func (r *ExamRepo) GetByIDs(ids []uint) ([]models.Exam, error) {
	var list []models.Exam
	err := r.db.Where("id IN ?", ids).Find(&list).Error
	return list, err
}

func (r *ExamRepo) ListByStatus(statuses []string) ([]models.Exam, error) {
	var list []models.Exam
	err := r.db.Where("status IN ?", statuses).Order("start_time asc").Find(&list).Error
	return list, err
}
