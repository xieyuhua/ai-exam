package repository

import (
	"exam-system/models"

	"gorm.io/gorm"
)

type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) GetAll() ([]models.Category, error) {
	var list []models.Category
	err := r.db.Order("id desc").Find(&list).Error
	return list, err
}

func (r *CategoryRepo) GetByID(id uint) (*models.Category, error) {
	var cat models.Category
	err := r.db.First(&cat, id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepo) Create(cat *models.Category) error {
	return r.db.Create(cat).Error
}

func (r *CategoryRepo) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Category{}).Where("id = ?", id).Updates(updates).Error
}

func (r *CategoryRepo) Delete(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}

func (r *CategoryRepo) CountExamsByCategory(categoryID uint) int64 {
	var count int64
	r.db.Model(&models.Exam{}).Where("category_id = ?", categoryID).Count(&count)
	return count
}

func (r *CategoryRepo) CountQuestionsByCategory(categoryID uint) int64 {
	var count int64
	r.db.Model(&models.Question{}).Where("category_id = ?", categoryID).Count(&count)
	return count
}

func (r *CategoryRepo) DeleteQuestionsByCategory(categoryID uint) error {
	return r.db.Where("category_id = ?", categoryID).Delete(&models.Question{}).Error
}
