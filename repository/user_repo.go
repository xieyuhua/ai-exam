package repository

import (
	"exam-system/models"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByWorkNo(workNo string) (*models.User, error) {
	var user models.User
	err := r.db.Where("work_no = ?", workNo).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepo) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *UserRepo) UpdatePassword(id uint, newHash string) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("password", newHash).Error
}

func (r *UserRepo) CountByWorkNo(workNo string) int64 {
	var count int64
	r.db.Model(&models.User{}).Where("work_no = ?", workNo).Count(&count)
	return count
}
