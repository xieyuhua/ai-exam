package service

import (
	"exam-system/models"
	"exam-system/repository"
	resp "exam-system/dto/response"
)

type ScoreService struct {
	repo     *repository.ScoreRepo
	examRepo *repository.ExamRepo
	catRepo  *repository.CategoryRepo
}

func NewScoreService(repo *repository.ScoreRepo, examRepo *repository.ExamRepo, catRepo *repository.CategoryRepo) *ScoreService {
	return &ScoreService{repo: repo, examRepo: examRepo, catRepo: catRepo}
}

func (s *ScoreService) List(keyword, categoryID string, page, pageSize int) (*resp.PagedData, error) {
	scores, total, err := s.repo.List(keyword, categoryID, page, pageSize)
	if err != nil {
		return nil, err
	}

	result := make([]resp.ScoreWithDetail, len(scores))
	for i, sc := range scores {
		result[i] = resp.ScoreWithDetail{Score: sc}
		if exam, err := s.examRepo.GetByID(sc.ExamID); err == nil {
			result[i].ExamTitle = exam.Title
			result[i].CanViewAnswer = true
			if exam.CanViewAnswer != nil {
				result[i].CanViewAnswer = *exam.CanViewAnswer
			}
			if cat, err := s.catRepo.GetByID(exam.CategoryID); err == nil {
				result[i].CategoryName = cat.Name
			}
		}
	}

	return &resp.PagedData{
		List:     result,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ScoreService) Save(score *models.Score) error {
	return s.repo.Create(score)
}

func (s *ScoreService) GetByExamAndStudent(examID uint, studentName string) (*models.Score, error) {
	return s.repo.GetByExamAndStudent(examID, studentName)
}

func (s *ScoreService) CountByExamAndStudent(examID uint, studentName string) int64 {
	return s.repo.CountByExamAndStudent(examID, studentName)
}

func (s *ScoreService) GetAllForExport(categoryID string) ([]models.Score, error) {
	return s.repo.GetAllForExport(categoryID)
}

func (s *ScoreService) ListByStudent(studentName string) ([]models.Score, error) {
	return s.repo.ListByStudent(studentName)
}

// ScoreExportRow 成绩导出数据结构
type ScoreExportRow struct {
	StudentName  string
	ExamTitle    string
	CategoryName string
	Date         string
	Score        float64
	Correct      int
	Total        int
}

func (s *ScoreService) BuildExportRowsByExam(examID uint) ([]ScoreExportRow, error) {
	scores, err := s.repo.GetByExamID(examID)
	if err != nil {
		return nil, err
	}

	exam, _ := s.examRepo.GetByID(examID)
	cat, _ := s.catRepo.GetByID(exam.CategoryID)

	catName := ""
	if cat != nil {
		catName = cat.Name
	}
	examTitle := ""
	if exam != nil {
		examTitle = exam.Title
	}

	var rows []ScoreExportRow
	for _, sc := range scores {
		rows = append(rows, ScoreExportRow{
			StudentName:  sc.StudentName,
			ExamTitle:    examTitle,
			CategoryName: catName,
			Date:         sc.CreatedAt.Local().Format("2006-01-02 15:04"),
			Score:        sc.Score,
			Correct:      sc.Correct,
			Total:        sc.Total,
		})
	}
	return rows, nil
}

func (s *ScoreService) BuildExportRows(categoryID string) ([]ScoreExportRow, error) {
	scores, err := s.repo.GetAllForExport(categoryID)
	if err != nil {
		return nil, err
	}

	var rows []ScoreExportRow
	for _, sc := range scores {
		exam, _ := s.examRepo.GetByID(sc.ExamID)
		cat, _ := s.catRepo.GetByID(exam.CategoryID)

		catName := ""
		if cat != nil {
			catName = cat.Name
		}
		examTitle := ""
		if exam != nil {
			examTitle = exam.Title
		}

		rows = append(rows, ScoreExportRow{
			StudentName:  sc.StudentName,
			ExamTitle:    examTitle,
			CategoryName: catName,
			Date:         sc.CreatedAt.Local().Format("2006-01-02 15:04"),
			Score:        sc.Score,
			Correct:      sc.Correct,
			Total:        sc.Total,
		})
	}
	return rows, nil
}
