package repository

import (
	"exam-system/models"

	"gorm.io/gorm"
)

type StudentRepo struct {
	db *gorm.DB
}

func NewStudentRepo(db *gorm.DB) *StudentRepo {
	return &StudentRepo{db: db}
}

func (r *StudentRepo) List(keyword string, page, pageSize int) ([]models.Student, int64, error) {
	query := r.db.Model(&models.Student{})
	if keyword != "" {
		query = query.Where("work_no LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []models.Student
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	// 计算 hasPassword
	for i := range list {
		list[i].HasPassword = list[i].Password != ""
	}
	return list, total, err
}

func (r *StudentRepo) GetByID(id uint) (*models.Student, error) {
	var student models.Student
	err := r.db.First(&student, id).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *StudentRepo) GetByWorkNo(workNo string) (*models.Student, error) {
	var student models.Student
	err := r.db.Where("work_no = ?", workNo).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *StudentRepo) GetByWxUserID(wxUserID string) (*models.Student, error) {
	var student models.Student
	err := r.db.Where("wx_user_id = ?", wxUserID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *StudentRepo) GetByWorkNoAndEmptyWx(workNo string) (*models.Student, error) {
	var student models.Student
	err := r.db.Where("work_no = ? AND wx_user_id = ''", workNo).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *StudentRepo) Create(student *models.Student) error {
	return r.db.Create(student).Error
}

func (r *StudentRepo) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Student{}).Where("id = ?", id).Updates(updates).Error
}

func (r *StudentRepo) Delete(id uint) error {
	return r.db.Delete(&models.Student{}, id).Error
}

func (r *StudentRepo) CountByWorkNo(workNo string) int64 {
	var count int64
	r.db.Model(&models.Student{}).Where("work_no = ?", workNo).Count(&count)
	return count
}

func (r *StudentRepo) ExistsByWorkNo(workNo string) bool {
	return r.CountByWorkNo(workNo) > 0
}
