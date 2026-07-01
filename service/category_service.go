package service

import (
	"exam-system/models"
	"exam-system/repository"
	resp "exam-system/dto/response"
	"errors"
)

type CategoryService struct {
	repo *repository.CategoryRepo
}

func NewCategoryService(repo *repository.CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) GetAll() ([]resp.CategoryWithCount, error) {
	categories, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	result := make([]resp.CategoryWithCount, len(categories))
	for i, cat := range categories {
		result[i] = resp.CategoryWithCount{
			Category:      cat,
			ExamCount:     int(s.repo.CountExamsByCategory(cat.ID)),
			QuestionCount: int(s.repo.CountQuestionsByCategory(cat.ID)),
		}
	}
	return result, nil
}

func (s *CategoryService) Create(name, desc string) (*models.Category, error) {
	cat := models.Category{Name: name, Desc: desc}
	if err := s.repo.Create(&cat); err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *CategoryService) Update(id uint, name, desc string) error {
	return s.repo.Update(id, map[string]interface{}{"name": name, "desc": desc})
}

func (s *CategoryService) Delete(id uint) error {
	if s.repo.CountExamsByCategory(id) > 0 {
		return errors.New("该分类下存在考试，请先删除考试")
	}
	s.repo.DeleteQuestionsByCategory(id)
	return s.repo.Delete(id)
}
