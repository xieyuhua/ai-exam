package service

import (
	"encoding/json"
	"exam-system/models"
	"exam-system/repository"
	req "exam-system/dto/request"
	resp "exam-system/dto/response"
	"errors"
	"sort"
)

type VoteService struct {
	repo *repository.VoteRepo
}

func NewVoteService(repo *repository.VoteRepo) *VoteService {
	return &VoteService{repo: repo}
}

func (s *VoteService) List(keyword, status string, page, pageSize int) (*resp.PagedData, error) {
	s.repo.AutoUpdateStatus()
	list, total, err := s.repo.List(keyword, status, page, pageSize)
	if err != nil {
		return nil, err
	}
	for i := range list {
		cnt, _ := s.repo.CountRecords(list[i].ID)
		list[i].TotalVotes = int(cnt)
	}
	return &resp.PagedData{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *VoteService) Create(r req.CreateVoteReq) (*models.Vote, error) {
	voteType := r.VoteType
	if voteType == "" {
		voteType = "single"
	}
	maxChoices := r.MaxChoices
	if maxChoices < 1 {
		maxChoices = 1
	}
	v := models.Vote{
		Title:       r.Title,
		Description: r.Description,
		VoteType:    voteType,
		MaxChoices:  maxChoices,
		AllowRepeat: r.AllowRepeat,
		IsPublic:    r.IsPublic,
		Status:      "upcoming",
	}
	if r.StartTime != "" {
		v.StartTime, _ = parseTime(r.StartTime)
	}
	if r.EndTime != "" {
		v.EndTime, _ = parseTime(r.EndTime)
	}

	if err := s.repo.Create(&v); err != nil {
		return nil, err
	}

	opts := make([]models.VoteOption, len(r.Options))
	for i, o := range r.Options {
		optType := o.Type
		if optType == "" {
			optType = "text"
		}
		opts[i] = models.VoteOption{
			VoteID:    v.ID,
			Label:     o.Label,
			Type:      optType,
			Content:   o.Content,
			URL:       o.URL,
			SortOrder: i,
		}
	}
	if err := s.repo.CreateOptions(opts); err != nil {
		return nil, err
	}

	return &v, nil
}

func (s *VoteService) Update(id uint, r req.CreateVoteReq) error {
	updates := map[string]interface{}{}
	if r.Title != "" {
		updates["title"] = r.Title
	}
	if r.Description != "" {
		updates["description"] = r.Description
	}
	if r.VoteType != "" {
		updates["vote_type"] = r.VoteType
	}
	if r.MaxChoices > 0 {
		updates["max_choices"] = r.MaxChoices
	}
	if r.AllowRepeat != nil {
		updates["allow_repeat"] = *r.AllowRepeat
	}
	if r.IsPublic != nil {
		updates["is_public"] = *r.IsPublic
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
		if err := s.repo.Update(id, updates); err != nil {
			return err
		}
	}
	if len(r.Options) > 0 {
		s.repo.DeleteOptions(id)
		opts := make([]models.VoteOption, len(r.Options))
		for i, o := range r.Options {
			optType := o.Type
			if optType == "" {
				optType = "text"
			}
			opts[i] = models.VoteOption{
				VoteID:    id,
				Label:     o.Label,
				Type:      optType,
				Content:   o.Content,
				URL:       o.URL,
				SortOrder: i,
			}
		}
		return s.repo.CreateOptions(opts)
	}
	return nil
}

func (s *VoteService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *VoteService) GetDetail(id uint, studentName string) (*resp.VoteWithDetail, error) {
	s.repo.AutoUpdateStatus()
	v, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	opts, err := s.repo.GetOptions(id)
	if err != nil {
		return nil, err
	}
	records, _ := s.repo.GetAllRecords(id)
	totalVotes := len(records)

	optStats := make(map[uint]int)
	for _, rec := range records {
		var oids []uint
		json.Unmarshal([]byte(rec.OptionIDs), &oids)
		for _, oid := range oids {
			optStats[oid]++
		}
	}

	hasVoted := false
	if studentName != "" {
		_, err := s.repo.GetRecord(id, studentName)
		hasVoted = (err == nil)
	}

	var optionStats []resp.VoteOptionStat
	for _, opt := range opts {
		cnt := optStats[opt.ID]
		pct := 0.0
		if totalVotes > 0 {
			pct = float64(cnt) / float64(totalVotes) * 100
		}
		optionStats = append(optionStats, resp.VoteOptionStat{
			VoteOption: opt,
			Count:      cnt,
			Percent:    pct,
		})
	}

	return &resp.VoteWithDetail{
		Vote:       *v,
		Options:    optionStats,
		TotalVotes: totalVotes,
		HasVoted:   hasVoted,
	}, nil
}

func (s *VoteService) Submit(studentName string, r req.SubmitVoteReq) error {
	s.repo.AutoUpdateStatus()
	v, err := s.repo.GetByID(r.VoteID)
	if err != nil {
		return errors.New("投票不存在")
	}
	if v.Status != "active" {
		return errors.New("投票未开放")
	}
	if v.AllowRepeat == nil || !*v.AllowRepeat {
		if _, err := s.repo.GetRecord(r.VoteID, studentName); err == nil {
			return errors.New("您已经投过票了")
		}
	}
	if v.VoteType == "single" && len(r.OptionIDs) > 1 {
		return errors.New("单选投票只能选一项")
	}
	if v.VoteType == "multiple" && len(r.OptionIDs) > v.MaxChoices {
		return errors.New("超过了最大可选数量")
	}

	optionIDsJSON, _ := json.Marshal(r.OptionIDs)
	record := models.VoteRecord{
		VoteID:      r.VoteID,
		StudentName: studentName,
		OptionIDs:   string(optionIDsJSON),
	}
	return s.repo.CreateRecord(&record)
}

func (s *VoteService) GetStudentVotes(studentName string) ([]resp.StudentVoteItem, error) {
	s.repo.AutoUpdateStatus()
	votes, _, _ := s.repo.List("", "", 1, 100)

	var items []resp.StudentVoteItem
	for _, v := range votes {
		hasVoted := false
		if _, err := s.repo.GetRecord(v.ID, studentName); err == nil {
			hasVoted = true
		}
		total, _ := s.repo.CountRecords(v.ID)

		items = append(items, resp.StudentVoteItem{
			ID:          v.ID,
			Title:       v.Title,
			Description: v.Description,
			VoteType:    v.VoteType,
			Status:      v.Status,
			StartTime:   v.StartTime.Format("2006-01-02 15:04"),
			EndTime:     v.EndTime.Format("2006-01-02 15:04"),
			AllowRepeat: v.AllowRepeat != nil && *v.AllowRepeat,
			IsPublic:    v.IsPublic == nil || *v.IsPublic,
			HasVoted:    hasVoted,
			TotalVotes:  int(total),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID > items[j].ID })
	return items, nil
}
