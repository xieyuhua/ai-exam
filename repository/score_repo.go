package repository

import (
	"exam-system/models"

	"gorm.io/gorm"
)

type ScoreRepo struct {
	db *gorm.DB
}

func NewScoreRepo(db *gorm.DB) *ScoreRepo {
	return &ScoreRepo{db: db}
}

func (r *ScoreRepo) List(keyword, categoryID string, page, pageSize int) ([]models.Score, int64, error) {
	query := r.db.Model(&models.Score{})
	if categoryID != "" {
		query = query.Where("exam_id IN (SELECT id FROM exams WHERE category_id = ?)", categoryID)
	}
	if keyword != "" {
		query = query.Joins("JOIN exams ON exams.id = scores.exam_id").
			Where("scores.student_name LIKE ? OR exams.title LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []models.Score
	err := query.Order("scores.id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *ScoreRepo) Create(score *models.Score) error {
	return r.db.Create(score).Error
}

func (r *ScoreRepo) GetByExamAndStudent(examID uint, studentName string) (*models.Score, error) {
	var score models.Score
	err := r.db.Where("exam_id = ? AND student_name = ?", examID, studentName).Order("id desc").First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

func (r *ScoreRepo) CountByExamAndStudent(examID uint, studentName string) int64 {
	var count int64
	r.db.Model(&models.Score{}).Where("exam_id = ? AND student_name = ?", examID, studentName).Count(&count)
	return count
}

func (r *ScoreRepo) GetAllForExport(categoryID string) ([]models.Score, error) {
	query := r.db.Model(&models.Score{})
	if categoryID != "" {
		query = query.Where("exam_id IN (SELECT id FROM exams WHERE category_id = ?)", categoryID)
	}
	var list []models.Score
	err := query.Order("id desc").Find(&list).Error
	return list, err
}

func (r *ScoreRepo) DeleteByExamID(tx *gorm.DB, examID uint) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Where("exam_id = ?", examID).Delete(&models.Score{}).Error
}

func (r *ScoreRepo) GetByExamID(examID uint) ([]models.Score, error) {
	var list []models.Score
	err := r.db.Where("exam_id = ?", examID).Order("id desc").Find(&list).Error
	return list, err
}

func (r *ScoreRepo) ListByStudent(studentName string) ([]models.Score, error) {
	query := r.db.Model(&models.Score{})
	if studentName != "" {
		query = query.Where("student_name = ?", studentName)
	}
	var list []models.Score
	err := query.Order("id desc").Find(&list).Error
	return list, err
}
