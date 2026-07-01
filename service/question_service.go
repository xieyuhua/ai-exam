package service

import (
	"encoding/json"
	"exam-system/dto"
	"exam-system/models"
	"exam-system/repository"
	req "exam-system/dto/request"
	resp "exam-system/dto/response"
	"exam-system/util"
	"errors"
	"sort"
)

type QuestionService struct {
	repo    *repository.QuestionRepo
	catRepo *repository.CategoryRepo
}

func NewQuestionService(repo *repository.QuestionRepo, catRepo *repository.CategoryRepo) *QuestionService {
	return &QuestionService{repo: repo, catRepo: catRepo}
}

func (s *QuestionService) List(categoryID, keyword, qType string, page, pageSize int) (*resp.PagedData, error) {
	list, total, err := s.repo.List(categoryID, keyword, qType, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &resp.PagedData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *QuestionService) Create(r req.CreateQuestionReq) (*models.Question, error) {
	qType := util.NormalizeQuestionType(r.Type)

	if qType == "judge" {
		r.Options = []dto.RichOption{
			{Label: "a", Type: "text", Content: "正确"},
			{Label: "b", Type: "text", Content: "错误"},
		}
	}
	if qType == "fill" || qType == "essay" {
		r.Options = nil
	}

	optionsJSON, _ := json.Marshal(r.Options)
	answerJSON, _ := json.Marshal(r.Answer)

	q := models.Question{
		CategoryID:  r.CategoryID,
		Type:        qType,
		Title:       r.Title,
		Options:     string(optionsJSON),
		Answer:      string(answerJSON),
		Explanation: r.Explanation,
	}
	if err := s.repo.Create(&q); err != nil {
		return nil, err
	}
	return &q, nil
}

func (s *QuestionService) Update(id uint, r req.UpdateQuestionReq) error {
	updates := map[string]interface{}{}
	if r.CategoryID != 0 {
		updates["category_id"] = r.CategoryID
	}
	if r.Type != "" {
		updates["type"] = util.NormalizeQuestionType(r.Type)
	}
	if r.Title != "" {
		updates["title"] = r.Title
	}
	if len(r.Options) > 0 {
		optionsJSON, _ := json.Marshal(r.Options)
		updates["options"] = string(optionsJSON)
	} else if r.Type != "" && util.NormalizeQuestionType(r.Type) == "judge" {
		opts := []dto.RichOption{
			{Label: "a", Type: "text", Content: "正确"},
			{Label: "b", Type: "text", Content: "错误"},
		}
		optionsJSON, _ := json.Marshal(opts)
		updates["options"] = string(optionsJSON)
	}
	if r.Type != "" && (util.NormalizeQuestionType(r.Type) == "fill" || util.NormalizeQuestionType(r.Type) == "essay") {
		updates["options"] = "null"
	}
	if r.Answer != nil {
		answerJSON, _ := json.Marshal(r.Answer)
		updates["answer"] = string(answerJSON)
	}
	if r.Explanation != "" {
		updates["explanation"] = r.Explanation
	}

	if len(updates) == 0 {
		return errors.New("没有需要更新的字段")
	}
	return s.repo.Update(id, updates)
}

func (s *QuestionService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *QuestionService) ImportBatch(categoryID uint, rows []ImportQuestionRow) (int, error) {
	count := 0
	for _, item := range rows {
		opts := ConvertOptions(item.Options)
		optionsJSON, _ := json.Marshal(opts)
		answerJSON, _ := json.Marshal(item.Answer)
		q := models.Question{
			CategoryID:  categoryID,
			Type:        util.NormalizeQuestionType(item.Type),
			Title:       item.Title,
			Options:     string(optionsJSON),
			Answer:      string(answerJSON),
			Explanation: item.Explanation,
		}
		if err := s.repo.Create(&q); err == nil {
			count++
		}
	}
	return count, nil
}

// ConvertOptions 兼容旧格式 map[string]string → []dto.RichOption
func ConvertOptions(m map[string]string) []dto.RichOption {
	if m == nil {
		return nil
	}
	var opts []dto.RichOption
	for k, v := range m {
		opts = append(opts, dto.RichOption{Label: k, Type: "text", Content: v})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

func (s *QuestionService) AvailableForExam(examID uint, keyword, qType string) ([]availQ, error) {
	exam, _ := s.catRepo.GetByID(examID)
	_ = exam
	return nil, nil
}

type availQ struct {
	ID      uint           `json:"id"`
	Type    string         `json:"type"`
	Title   string         `json:"title"`
	Options []dto.OptionPair `json:"options"`
}

// ParseOptions 解析选项 JSON（支持富媒体和旧格式）
func ParseOptions(optionsJSON string) []dto.OptionPair {
	opts := make([]dto.OptionPair, 0)
	if optionsJSON == "" || optionsJSON == "null" {
		return opts
	}
	var richOpts []dto.RichOption
	if err := json.Unmarshal([]byte(optionsJSON), &richOpts); err == nil && len(richOpts) > 0 {
		for _, ro := range richOpts {
			opts = append(opts, dto.OptionPair{
				Label:   ro.Label,
				Type:    ro.Type,
				Content: ro.Content,
				URL:     ro.URL,
			})
		}
		sort.Slice(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
		return opts
	}
	var optsMap map[string]string
	if err := json.Unmarshal([]byte(optionsJSON), &optsMap); err == nil {
		for k, v := range optsMap {
			opts = append(opts, dto.OptionPair{Label: k, Type: "text", Content: v})
		}
		sort.Slice(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	}
	return opts
}

// ParseAnswer 解析答案 JSON
func ParseAnswer(answerJSON string) []string {
	var ans []string
	json.Unmarshal([]byte(answerJSON), &ans)
	return ans
}
