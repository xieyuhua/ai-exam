package models

import "time"

// Vote 投票主题表
type Vote struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	VoteType    string    `gorm:"size:20;default:single" json:"voteType"`
	MaxChoices  int       `gorm:"default:1" json:"maxChoices"`
	AllowRepeat *bool     `gorm:"default:false" json:"allowRepeat"`
	IsPublic    *bool     `gorm:"default:true" json:"isPublic"`
	Status      string    `gorm:"size:20;default:upcoming" json:"status"`
	TotalVotes  int       `gorm:"-" json:"totalVotes"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// VoteOption 投票选项表
type VoteOption struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	VoteID    uint      `gorm:"index;not null" json:"voteId"`
	Label     string    `gorm:"size:10;not null" json:"label"`
	Type      string    `gorm:"size:20;default:text" json:"type"`
	Content   string    `gorm:"type:text" json:"content"`
	URL       string    `gorm:"size:500" json:"url"`
	SortOrder int       `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
}

// VoteRecord 投票记录表
type VoteRecord struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	VoteID      uint      `gorm:"index;not null" json:"voteId"`
	StudentName string    `gorm:"size:100;not null" json:"studentName"`
	OptionIDs   string    `gorm:"size:200;not null" json:"optionIds"`
	CreatedAt   time.Time `json:"createdAt"`
}
