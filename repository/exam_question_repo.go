package repository

import (
	"exam-system/models"

	"gorm.io/gorm"
)

type ExamQuestionRepo struct {
	db *gorm.DB
}

func NewExamQuestionRepo(db *gorm.DB) *ExamQuestionRepo {
	return &ExamQuestionRepo{db: db}
}

func (r *ExamQuestionRepo) GetByExamID(examID uint) ([]models.ExamQuestion, error) {
	var list []models.ExamQuestion
	err := r.db.Where("exam_id = ?", examID).Order("question_id asc").Find(&list).Error
	return list, err
}

func (r *ExamQuestionRepo) Create(eq *models.ExamQuestion) error {
	return r.db.Create(eq).Error
}

func (r *ExamQuestionRepo) CreateBatch(tx *gorm.DB, examID uint, questionIDs []uint) error {
	for _, qid := range questionIDs {
		if err := tx.Create(&models.ExamQuestion{ExamID: examID, QuestionID: qid}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *ExamQuestionRepo) DeleteByExamID(tx *gorm.DB, examID uint) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Where("exam_id = ?", examID).Delete(&models.ExamQuestion{}).Error
}

func (r *ExamQuestionRepo) DeleteByExamAndQuestion(examID, questionID uint) error {
	return r.db.Where("exam_id = ? AND question_id = ?", examID, questionID).Delete(&models.ExamQuestion{}).Error
}

func (r *ExamQuestionRepo) UpdateScore(examID, questionID uint, score float64) (int64, error) {
	result := r.db.Model(&models.ExamQuestion{}).
		Where("exam_id = ? AND question_id = ?", examID, questionID).
		Update("score", score)
	return result.RowsAffected, result.Error
}

func (r *ExamQuestionRepo) CountByExamID(examID uint) int64 {
	var count int64
	r.db.Model(&models.ExamQuestion{}).Where("exam_id = ?", examID).Count(&count)
	return count
}

func (r *ExamQuestionRepo) Exists(examID, questionID uint) bool {
	var count int64
	r.db.Model(&models.ExamQuestion{}).Where("exam_id = ? AND question_id = ?", examID, questionID).Count(&count)
	return count > 0
}

func (r *ExamQuestionRepo) GetExistingQuestionIDs(examID uint) ([]uint, error) {
	var eqs []models.ExamQuestion
	if err := r.db.Where("exam_id = ?", examID).Find(&eqs).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, len(eqs))
	for i, eq := range eqs {
		ids[i] = eq.QuestionID
	}
	return ids, nil
}
