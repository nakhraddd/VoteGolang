package data

import "gorm.io/gorm"

type PetitionVote struct {
	gorm.Model
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	UserID     string `json:"user_id"`
	PetitionID uint   `json:"petition_id"`
	VoteType   string `json:"vote_type"`
}
