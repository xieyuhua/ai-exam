package request

import "exam-system/dto"

// CreateVoteReq 创建投票请求
type CreateVoteReq struct {
	Title       string         `json:"title" binding:"required"`
	Description string         `json:"description"`
	StartTime   string         `json:"startTime" binding:"required"`
	EndTime     string         `json:"endTime" binding:"required"`
	VoteType    string         `json:"voteType"`
	MaxChoices  int            `json:"maxChoices"`
	AllowRepeat *bool          `json:"allowRepeat"`
	IsPublic    *bool          `json:"isPublic"`
	Options     []dto.RichOption `json:"options" binding:"required"`
}

// SubmitVoteReq 提交投票请求
type SubmitVoteReq struct {
	VoteID    uint   `json:"voteId" binding:"required"`
	OptionIDs []uint `json:"optionIds" binding:"required"`
}
