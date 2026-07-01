package repository

import (
	"exam-system/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type VoteRepo struct {
	db *gorm.DB
}

func NewVoteRepo(db *gorm.DB) *VoteRepo {
	return &VoteRepo{db: db}
}

func (r *VoteRepo) List(keyword, status string, page, pageSize int) ([]models.Vote, int64, error) {
	query := r.db.Model(&models.Vote{})
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	var list []models.Vote
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *VoteRepo) GetByID(id uint) (*models.Vote, error) {
	var v models.Vote
	err := r.db.First(&v, id).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VoteRepo) Create(v *models.Vote) error {
	return r.db.Create(v).Error
}

func (r *VoteRepo) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Vote{}).Where("id = ?", id).Updates(updates).Error
}

func (r *VoteRepo) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("vote_id = ?", id).Delete(&models.VoteOption{}).Error; err != nil {
			return err
		}
		if err := tx.Where("vote_id = ?", id).Delete(&models.VoteRecord{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Vote{}, id).Error
	})
}

func (r *VoteRepo) UpdateStatus(ids []uint, status string) error {
	return r.db.Model(&models.Vote{}).Where("id IN ?", ids).Update("status", status).Error
}

func (r *VoteRepo) AutoUpdateStatus() {
	// SQLite 不存时区，传格式化字符串（无时区后缀）确保字符串比较一致
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	r.db.Model(&models.Vote{}).Where("status = ? AND start_time <= ? AND end_time >= ?", "upcoming", nowStr, nowStr).Update("status", "active")
	r.db.Model(&models.Vote{}).Where("status = ? AND end_time < ?", "active", nowStr).Update("status", "ended")
}

// VoteOption
func (r *VoteRepo) CreateOptions(options []models.VoteOption) error {
	return r.db.Create(&options).Error
}

func (r *VoteRepo) GetOptions(voteID uint) ([]models.VoteOption, error) {
	var opts []models.VoteOption
	err := r.db.Where("vote_id = ?", voteID).Order("sort_order asc, id asc").Find(&opts).Error
	return opts, err
}

func (r *VoteRepo) DeleteOptions(voteID uint) error {
	return r.db.Where("vote_id = ?", voteID).Delete(&models.VoteOption{}).Error
}

// VoteRecord
func (r *VoteRepo) CreateRecord(record *models.VoteRecord) error {
	return r.db.Create(record).Error
}

func (r *VoteRepo) GetRecord(voteID uint, studentName string) (*models.VoteRecord, error) {
	var rec models.VoteRecord
	err := r.db.Where("vote_id = ? AND student_name = ?", voteID, studentName).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *VoteRepo) CountRecords(voteID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.VoteRecord{}).Where("vote_id = ?", voteID).Count(&count).Error
	return count, err
}

func (r *VoteRepo) CountOptionRecords(optionID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.VoteRecord{}).Where("option_ids LIKE ?", "%"+fmt.Sprint(optionID)+"%").Count(&count).Error
	return count, err
}

func (r *VoteRepo) GetAllRecords(voteID uint) ([]models.VoteRecord, error) {
	var records []models.VoteRecord
	err := r.db.Where("vote_id = ?", voteID).Find(&records).Error
	return records, err
}

func (r *VoteRepo) CountTotal() (int64, error) {
	var count int64
	err := r.db.Model(&models.Vote{}).Count(&count).Error
	return count, err
}

func (r *VoteRepo) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&models.Vote{}).Where("status = ?", "active").Count(&count).Error
	return count, err
}
