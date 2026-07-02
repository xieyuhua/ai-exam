package service

import (
	"exam-system/models"
	"exam-system/repository"
	req "exam-system/dto/request"
	resp "exam-system/dto/response"
	"time"

	"gorm.io/gorm"
)

type ExamService struct {
	repo    *repository.ExamRepo
	eqRepo  *repository.ExamQuestionRepo
	scRepo  *repository.ScoreRepo
	catRepo *repository.CategoryRepo
	qRepo   *repository.QuestionRepo
}

func NewExamService(repo *repository.ExamRepo, eqRepo *repository.ExamQuestionRepo, scRepo *repository.ScoreRepo, catRepo *repository.CategoryRepo, qRepo *repository.QuestionRepo) *ExamService {
	return &ExamService{repo: repo, eqRepo: eqRepo, scRepo: scRepo, catRepo: catRepo, qRepo: qRepo}
}

func (s *ExamService) List(categoryID, keyword, status string, page, pageSize int) (*resp.PagedData, error) {
	exams, total, err := s.repo.List(categoryID, keyword, status, page, pageSize)
	if err != nil {
		return nil, err
	}

	type examDetail struct {
		models.Exam
		CategoryName  string  `json:"categoryName"`
		QuestionCount int     `json:"questionCount"`
		ActualScore   float64 `json:"actualScore"`
		QuestionIDs   []uint  `json:"questionIds"`
	}

	now := time.Now()
	loc := now.Location()
	result := make([]examDetail, len(exams))
	for i, e := range exams {
		// DB(SQLite) 不存时区，按本地时区重新解释后实时计算状态
		startTime := time.Date(e.StartTime.Year(), e.StartTime.Month(), e.StartTime.Day(),
			e.StartTime.Hour(), e.StartTime.Minute(), e.StartTime.Second(), 0, loc)
		endTime := time.Date(e.EndTime.Year(), e.EndTime.Month(), e.EndTime.Day(),
			e.EndTime.Hour(), e.EndTime.Minute(), e.EndTime.Second(), 0, loc)
		e.StartTime = startTime
		e.EndTime = endTime
		if now.After(startTime) && now.Before(endTime) {
			e.Status = "active"
		} else if now.After(endTime) {
			e.Status = "ended"
		} else {
			e.Status = "upcoming"
		}

		cat, _ := s.catRepo.GetByID(e.CategoryID)
		qCount := s.eqRepo.CountByExamID(e.ID)
		eqs, _ := s.eqRepo.GetByExamID(e.ID)
		qids := make([]uint, 0, len(eqs))
		actualScore := 0.0
		for _, eq := range eqs {
			qids = append(qids, eq.QuestionID)
			actualScore += eq.Score
		}

		categoryName := ""
		if cat != nil {
			categoryName = cat.Name
		}
		result[i] = examDetail{
			Exam:          e,
			CategoryName:  categoryName,
			QuestionCount: int(qCount),
			ActualScore:   actualScore,
			QuestionIDs:   qids,
		}
	}

	return &resp.PagedData{
		List:     result,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ExamService) Create(r req.CreateExamReq, db *gorm.DB) (*models.Exam, error) {
	startTime, _ := time.ParseInLocation("2006-01-02T15:04", r.StartTime, time.Local)
	endTime, _ := time.ParseInLocation("2006-01-02T15:04", r.EndTime, time.Local)

	canView := true
	if r.CanViewAnswer != nil {
		canView = *r.CanViewAnswer
	}
	allowRepeat := false
	if r.AllowRepeat != nil {
		allowRepeat = *r.AllowRepeat
	}

	now := time.Now()
	status := "upcoming"
	if now.After(startTime) && now.Before(endTime) {
		status = "active"
	} else if now.After(endTime) {
		status = "ended"
	}

	exam := models.Exam{
		CategoryID:    r.CategoryID,
		Title:         r.Title,
		Duration:      r.Duration,
		TotalScore:    r.TotalScore,
		StartTime:     startTime,
		EndTime:       endTime,
		CanViewAnswer: &canView,
		AllowRepeat:   &allowRepeat,
		Status:        status,
	}

	tx := db.Begin()
	if err := s.repo.CreateWithTx(tx, &exam); err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(r.QuestionIDs) > 0 {
		if err := s.eqRepo.CreateBatch(tx, exam.ID, r.QuestionIDs); err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	tx.Commit()

	return &exam, nil
}

func (s *ExamService) Update(id uint, r req.UpdateExamReq) error {
	updates := map[string]interface{}{}
	if r.CategoryID != 0 {
		updates["category_id"] = r.CategoryID
	}
	if r.Title != "" {
		updates["title"] = r.Title
	}
	if r.Duration > 0 {
		updates["duration"] = r.Duration
	}
	if r.TotalScore > 0 {
		updates["total_score"] = r.TotalScore
	}
	if r.StartTime != "" {
		t, _ := time.ParseInLocation("2006-01-02T15:04", r.StartTime, time.Local)
		updates["start_time"] = t
	}
	if r.EndTime != "" {
		t, _ := time.ParseInLocation("2006-01-02T15:04", r.EndTime, time.Local)
		updates["end_time"] = t
	}
	if r.CanViewAnswer != nil {
		updates["can_view_answer"] = *r.CanViewAnswer
	}
	if r.AllowRepeat != nil {
		updates["allow_repeat"] = *r.AllowRepeat
	}

	if len(updates) > 0 {
		if err := s.repo.Update(id, updates); err != nil {
			return err
		}
		exam, _ := s.repo.GetByID(id)
		if exam != nil {
			now := time.Now()
			loc := now.Location()
			startTime := time.Date(exam.StartTime.Year(), exam.StartTime.Month(), exam.StartTime.Day(),
				exam.StartTime.Hour(), exam.StartTime.Minute(), exam.StartTime.Second(), 0, loc)
			endTime := time.Date(exam.EndTime.Year(), exam.EndTime.Month(), exam.EndTime.Day(),
				exam.EndTime.Hour(), exam.EndTime.Minute(), exam.EndTime.Second(), 0, loc)
			newStatus := "upcoming"
			if now.After(startTime) && now.Before(endTime) {
				newStatus = "active"
			} else if now.After(endTime) {
				newStatus = "ended"
			}
			s.repo.Update(id, map[string]interface{}{"status": newStatus})
		}
	}

	if r.QuestionIDs != nil {
		s.eqRepo.DeleteByExamID(nil, id)
		for _, qid := range r.QuestionIDs {
			s.eqRepo.Create(&models.ExamQuestion{ExamID: id, QuestionID: qid})
		}
	}

	return nil
}

func (s *ExamService) Delete(id uint) error {
	s.eqRepo.DeleteByExamID(nil, id)
	s.scRepo.DeleteByExamID(nil, id)
	return s.repo.Delete(id)
}

func (s *ExamService) GetByID(id uint) (*models.Exam, error) {
	return s.repo.GetByID(id)
}

func (s *ExamService) AddExamQuestions(examID uint, questionIDs []uint, defaultScore float64, scores map[uint]float64) (added, skipped int, err error) {
	for _, qid := range questionIDs {
		if s.eqRepo.Exists(examID, qid) {
			skipped++
			continue
		}
		score := defaultScore
		if scores != nil {
			if sc, ok := scores[qid]; ok {
				score = sc
			}
		}
		s.eqRepo.Create(&models.ExamQuestion{
			ExamID:     examID,
			QuestionID: qid,
			Score:      score,
		})
		added++
	}
	return
}

func (s *ExamService) UpdateExamQuestionScore(examID, questionID uint, score float64) (int64, error) {
	return s.eqRepo.UpdateScore(examID, questionID, score)
}

func (s *ExamService) RemoveExamQuestion(examID, questionID uint) error {
	return s.eqRepo.DeleteByExamAndQuestion(examID, questionID)
}

func (s *ExamService) ClearExamQuestions(examID uint) error {
	return s.eqRepo.DeleteByExamID(nil, examID)
}

func (s *ExamService) GetAvailableQuestions(categoryID uint, existingIDs []uint, keyword, qType string, limit int) ([]models.Question, error) {
	return s.qRepo.AvailableForExam(categoryID, existingIDs, keyword, qType, limit)
}
