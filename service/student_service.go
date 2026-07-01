package service

import (
	"exam-system/models"
	"exam-system/repository"
	req "exam-system/dto/request"
	resp "exam-system/dto/response"
	"exam-system/util"
	"errors"
	"fmt"
)

type StudentService struct {
	repo     *repository.StudentRepo
	userRepo *repository.UserRepo
}

func NewStudentService(repo *repository.StudentRepo, userRepo *repository.UserRepo) *StudentService {
	return &StudentService{repo: repo, userRepo: userRepo}
}

func (s *StudentService) List(keyword string, page, pageSize int) (*resp.PagedData, error) {
	students, total, err := s.repo.List(keyword, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &resp.PagedData{
		List:     students,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *StudentService) Create(workNo, name, password string) (*models.Student, error) {
	if s.repo.ExistsByWorkNo(workNo) {
		return nil, errors.New("学号 " + workNo + " 已存在")
	}
	if s.userRepo.CountByWorkNo(workNo) > 0 {
		return nil, errors.New("账号 " + workNo + " 已被管理员占用")
	}

	student := models.Student{
		WorkNo:   workNo,
		Name:     name,
		Password: util.HashPassword(password),
		Source:   "import",
	}
	if err := s.repo.Create(&student); err != nil {
		return nil, err
	}
	return &student, nil
}

func (s *StudentService) BatchCreate(students []req.StudentImportReq) (added, skipped int) {
	for _, st := range students {
		if st.WorkNo == "" || st.Name == "" || st.Password == "" {
			skipped++
			continue
		}
		if s.repo.ExistsByWorkNo(st.WorkNo) || s.userRepo.CountByWorkNo(st.WorkNo) > 0 {
			skipped++
			continue
		}
		student := models.Student{
			WorkNo:   st.WorkNo,
			Name:     st.Name,
			Password: util.HashPassword(st.Password),
			Source:   "import",
		}
		if err := s.repo.Create(&student); err == nil {
			added++
		} else {
			skipped++
		}
	}
	return
}

func (s *StudentService) Update(id uint, name, password string) error {
	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if password != "" {
		updates["password"] = util.HashPassword(password)
	}
	if len(updates) == 0 {
		return fmt.Errorf("没有需要更新的字段")
	}
	return s.repo.Update(id, updates)
}

func (s *StudentService) Delete(id uint) error {
	return s.repo.Delete(id)
}
