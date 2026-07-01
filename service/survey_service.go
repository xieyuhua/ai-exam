package service

import (
	"encoding/json"
	"exam-system/dto"
	"exam-system/models"
	"exam-system/repository"
	req "exam-system/dto/request"
	resp "exam-system/dto/response"
	"errors"
	"sort"
	"strings"
	"time"
)

type SurveyService struct {
	repo *repository.SurveyRepo
}

func NewSurveyService(repo *repository.SurveyRepo) *SurveyService {
	return &SurveyService{repo: repo}
}

func (s *SurveyService) List(keyword, status string, page, pageSize int) (*resp.PagedData, error) {
	s.repo.AutoUpdateStatus()
	list, total, err := s.repo.List(keyword, status, page, pageSize)
	if err != nil {
		return nil, err
	}
	for i := range list {
		questions, _ := s.repo.GetQuestions(list[i].ID)
		list[i].QuestionCount = len(questions)
		completed, _ := s.repo.CountCompleted(list[i].ID)
		list[i].TotalCompleted = completed
	}
	return &resp.PagedData{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *SurveyService) Create(r req.CreateSurveyReq) (*models.Survey, error) {
	survey := models.Survey{
		Title:       r.Title,
		Description: r.Description,
		AllowRepeat: r.AllowRepeat,
		Status:      "upcoming",
	}
	if r.StartTime != "" {
		survey.StartTime, _ = parseTime(r.StartTime)
	}
	if r.EndTime != "" {
		survey.EndTime, _ = parseTime(r.EndTime)
	}

	if err := s.repo.Create(&survey); err != nil {
		return nil, err
	}

	var questions []models.SurveyQuestion
	var allOptions []models.SurveyOption
	for i, qr := range r.Questions {
		qType := qr.Type
		if qType == "" {
			qType = "single"
		}
		required := true
		if qr.Required != nil {
			required = *qr.Required
		}
		order := qr.SortOrder
		if order == 0 {
			order = i
		}
		questions = append(questions, models.SurveyQuestion{
			SurveyID:  survey.ID,
			Title:     qr.Title,
			Type:      qType,
			SortOrder: order,
			Required:  &required,
		})
	}
	if err := s.repo.CreateQuestions(questions); err != nil {
		return nil, err
	}

	for qIdx, q := range questions {
		qReq := r.Questions[qIdx]
		for oIdx, o := range qReq.Options {
			optType := o.Type
			if optType == "" {
				optType = "text"
			}
			allOptions = append(allOptions, models.SurveyOption{
				SurveyQuestionID: q.ID,
				Label:            o.Label,
				Type:             optType,
				Content:          o.Content,
				URL:              o.URL,
				SortOrder:        oIdx,
			})
		}
	}
	if len(allOptions) > 0 {
		s.repo.CreateOptions(allOptions)
	}

	return &survey, nil
}

func (s *SurveyService) Update(id uint, r req.CreateSurveyReq) error {
	updates := map[string]interface{}{}
	if r.Title != "" {
		updates["title"] = r.Title
	}
	if r.Description != "" {
		updates["description"] = r.Description
	}
	if r.AllowRepeat != nil {
		updates["allow_repeat"] = *r.AllowRepeat
	}
	if r.StartTime != "" {
		t, err := parseTime(r.StartTime)
		if err == nil {
			updates["start_time"] = t
		}
	}
	if r.EndTime != "" {
		t, err := parseTime(r.EndTime)
		if err == nil {
			updates["end_time"] = t
		}
	}
	if len(updates) > 0 {
		s.repo.Update(id, updates)
	}

	if len(r.Questions) > 0 {
		oldQuestions, _ := s.repo.GetQuestions(id)
		var oldQIDs []uint
		for _, q := range oldQuestions {
			oldQIDs = append(oldQIDs, q.ID)
		}
		if len(oldQIDs) > 0 {
			s.repo.DeleteOptionsByQIDs(oldQIDs)
			s.repo.DeleteQuestions(id)
		}
		var newQuestions []models.SurveyQuestion
		var allOptions []models.SurveyOption
		for i, qr := range r.Questions {
			qType := qr.Type
			if qType == "" {
				qType = "single"
			}
			required := true
			if qr.Required != nil {
				required = *qr.Required
			}
			order := qr.SortOrder
			if order == 0 {
				order = i
			}
			newQuestions = append(newQuestions, models.SurveyQuestion{
				SurveyID:  id,
				Title:     qr.Title,
				Type:      qType,
				SortOrder: order,
				Required:  &required,
			})
		}
		s.repo.CreateQuestions(newQuestions)
		for qIdx, q := range newQuestions {
			qReq := r.Questions[qIdx]
			for oIdx, o := range qReq.Options {
				optType := o.Type
				if optType == "" {
					optType = "text"
				}
				allOptions = append(allOptions, models.SurveyOption{
					SurveyQuestionID: q.ID,
					Label:            o.Label,
					Type:             optType,
					Content:          o.Content,
					URL:              o.URL,
					SortOrder:        oIdx,
				})
			}
		}
		if len(allOptions) > 0 {
			s.repo.CreateOptions(allOptions)
		}
	}
	return nil
}

func (s *SurveyService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *SurveyService) GetDetail(id uint) (*resp.SurveyWithDetail, error) {
	s.repo.AutoUpdateStatus()
	survey, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	questions, err := s.repo.GetQuestions(id)
	if err != nil {
		return nil, err
	}
	var qIDs []uint
	for _, q := range questions {
		qIDs = append(qIDs, q.ID)
	}
	var allOpts []models.SurveyOption
	if len(qIDs) > 0 {
		allOpts, _ = s.repo.GetOptions(qIDs)
	}
	optMap := make(map[uint][]models.SurveyOption)
	for _, o := range allOpts {
		optMap[o.SurveyQuestionID] = append(optMap[o.SurveyQuestionID], o)
	}

	var qDetails []resp.SurveyQuestionDetail
	for _, q := range questions {
		opts := optMap[q.ID]
		if opts == nil {
			opts = []models.SurveyOption{}
		}
		var richOpts []dto.RichOption
		for _, o := range opts {
			richOpts = append(richOpts, dto.RichOption{
				Label:   o.Label,
				Type:    o.Type,
				Content: o.Content,
				URL:     o.URL,
			})
		}
		qDetails = append(qDetails, resp.SurveyQuestionDetail{
			SurveyQuestion: q,
			Options:        richOpts,
		})
	}

	return &resp.SurveyWithDetail{
		Survey:    *survey,
		Questions: qDetails,
	}, nil
}

func (s *SurveyService) Submit(studentName string, r req.SubmitSurveyReq) error {
	s.repo.AutoUpdateStatus()
	survey, err := s.repo.GetByID(r.SurveyID)
	if err != nil {
		return errors.New("问卷不存在")
	}
	if survey.Status != "active" {
		return errors.New("问卷未开放")
	}
	now := time.Now()
	loc := now.Location()
	startTime := time.Date(survey.StartTime.Year(), survey.StartTime.Month(), survey.StartTime.Day(),
		survey.StartTime.Hour(), survey.StartTime.Minute(), survey.StartTime.Second(), 0, loc)
	endTime := time.Date(survey.EndTime.Year(), survey.EndTime.Month(), survey.EndTime.Day(),
		survey.EndTime.Hour(), survey.EndTime.Minute(), survey.EndTime.Second(), 0, loc)
	if now.Before(startTime) || now.After(endTime) {
		return errors.New("不在问卷有效时间范围内")
	}

	if survey.AllowRepeat == nil || !*survey.AllowRepeat {
		existing, _ := s.repo.GetAnswer(r.SurveyID, studentName)
		if len(existing) > 0 {
			return errors.New("您已经提交过该问卷")
		}
	}

	var answers []models.SurveyAnswer
	for _, a := range r.Answers {
		ansJSON, _ := json.Marshal(a.Answer)
		answers = append(answers, models.SurveyAnswer{
			SurveyID:         r.SurveyID,
			SurveyQuestionID: a.SurveyQuestionID,
			StudentName:      studentName,
			Answer:           string(ansJSON),
		})
	}
	return s.repo.CreateAnswers(answers)
}

func (s *SurveyService) GetStudentSurveys(studentName string) ([]resp.StudentSurveyItem, error) {
	s.repo.AutoUpdateStatus()
	surveys, _, _ := s.repo.List("", "", 1, 100)
	var items []resp.StudentSurveyItem
	for _, survey := range surveys {
		questions, _ := s.repo.GetQuestions(survey.ID)
		hasCompleted := false
		answers, _ := s.repo.GetAnswer(survey.ID, studentName)
		if len(answers) > 0 {
			hasCompleted = true
		}
		items = append(items, resp.StudentSurveyItem{
			ID:            survey.ID,
			Title:         survey.Title,
			Description:   survey.Description,
			Status:        survey.Status,
			StartTime:     survey.StartTime.Format("2006-01-02 15:04"),
			EndTime:       survey.EndTime.Format("2006-01-02 15:04"),
			AllowRepeat:   survey.AllowRepeat != nil && *survey.AllowRepeat,
			HasCompleted:  hasCompleted,
			QuestionCount: len(questions),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID > items[j].ID })
	return items, nil
}

func (s *SurveyService) GetStatistics(id uint) (*resp.SurveyStatistics, error) {
	questions, _ := s.repo.GetQuestions(id)
	answers, _ := s.repo.GetAllAnswers(id)

	studentSet := make(map[string]bool)
	for _, a := range answers {
		studentSet[a.StudentName] = true
	}
	totalCompleted := int64(len(studentSet))

	var qIDs []uint
	for _, q := range questions {
		qIDs = append(qIDs, q.ID)
	}
	var allOpts []models.SurveyOption
	if len(qIDs) > 0 {
		allOpts, _ = s.repo.GetOptions(qIDs)
	}
	optMap := make(map[uint][]models.SurveyOption)
	for _, o := range allOpts {
		optMap[o.SurveyQuestionID] = append(optMap[o.SurveyQuestionID], o)
	}

	ansMap := make(map[uint][]models.SurveyAnswer)
	for _, a := range answers {
		ansMap[a.SurveyQuestionID] = append(ansMap[a.SurveyQuestionID], a)
	}

	var qStats []resp.SurveyQuestionStat
	for _, q := range questions {
		qAns := ansMap[q.ID]
		respCount := 0
		optCounts := make(map[string]int)
		var textResps []string

		for _, a := range qAns {
			respCount++
			var parsed []string
			json.Unmarshal([]byte(a.Answer), &parsed)
			if q.Type == "single" || q.Type == "multiple" {
				for _, label := range parsed {
					optCounts[label]++
				}
			} else if len(textResps) < 20 {
				textResps = append(textResps, strings.Join(parsed, ", "))
			}
		}

		stat := resp.SurveyQuestionStat{
			ID:            q.ID,
			Title:         q.Title,
			Type:          q.Type,
			Required:      q.Required != nil && *q.Required,
			ResponseCount: respCount,
			Options:       []resp.SurveyOptionStatItem{},
			TextResponses: textResps,
		}

		for _, o := range optMap[q.ID] {
			cnt := optCounts[o.Label]
			pct := 0.0
			if totalCompleted > 0 {
				pct = float64(cnt) / float64(totalCompleted) * 100
			}
			stat.Options = append(stat.Options, resp.SurveyOptionStatItem{
				Label:   o.Label,
				Type:    o.Type,
				Content: o.Content,
				URL:     o.URL,
				Count:   cnt,
				Percent: pct,
			})
		}

		qStats = append(qStats, stat)
	}

	return &resp.SurveyStatistics{
		SurveyID:       id,
		TotalCompleted: totalCompleted,
		TotalResponses: int64(len(answers)),
		Questions:      qStats,
	}, nil
}
