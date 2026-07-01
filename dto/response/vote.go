package response

import "exam-system/models"

// VoteWithDetail 投票详情（含选项和统计）
type VoteWithDetail struct {
	models.Vote
	Options    []VoteOptionStat `json:"options"`
	TotalVotes int              `json:"totalVotes"`
	HasVoted   bool             `json:"hasVoted"`
}

// VoteOptionStat 选项统计
type VoteOptionStat struct {
	models.VoteOption
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

// StudentVoteItem 员工端投票列表项
type StudentVoteItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	VoteType    string `json:"voteType"`
	Status      string `json:"status"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	AllowRepeat bool   `json:"allowRepeat"`
	IsPublic    bool   `json:"isPublic"`
	HasVoted    bool   `json:"hasVoted"`
	TotalVotes  int    `json:"totalVotes"`
}
