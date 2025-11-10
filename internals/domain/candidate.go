package domain

import (
	"time"

	"gorm.io/gorm"
)

// Candidate represents a candidate for election.
type Candidate struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" swaggerignore:"true"`
	Name           string         `gorm:"type:varchar(255);not null" example:"Beksultan"`
	Photo          *string        `gorm:"type:varchar(255)" example:"link"`
	Education      *string        `gorm:"type:varchar(255)" example:"KBTU"`
	Age            int            `gorm:"not null" example:"20"`
	Party          *string        `gorm:"type:varchar(255)" example:"Jastar"`
	Region         *string        `gorm:"type:varchar(255)" example:"SKO"`
	Votes          int            `gorm:"default:0" swaggerignore:"true"`
	Type           CandidateType  `gorm:"type:varchar(255);not null" example:"manager"`
	VotingStart    time.Time      `json:"voting_start" gorm:"type:datetime" example:"2025-11-12T09:00:00+05:00"`
	VotingDeadline time.Time      `json:"voting_deadline" gorm:"type:datetime" example:"2026-11-12T09:00:00+05:00"`
	DeletedAt      gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" swaggerignore:"true"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" swaggerignore:"true"`
}

// CandidateRepository retrieves candidate data.
type CandidateRepository interface {
	Create(candidate *Candidate) error
	GetAllByType(candidateType string) ([]Candidate, error)
	GetByID(id uint) (*Candidate, error)
	IncrementVote(id uint) error
	GetAllByTypePaginated(candidateType string, limit, offset int) ([]Candidate, error)
	DeleteByID(id uint) error
}

type CandidateType string

const (
	Presidential CandidateType = "presidential"
	Deputy       CandidateType = "deputy"
	Manager      CandidateType = "manager"
)

func IsValidCandidateType(t string) bool {
	switch CandidateType(t) {
	case Presidential, Deputy, Manager:
		return true
	default:
		return false
	}
}
