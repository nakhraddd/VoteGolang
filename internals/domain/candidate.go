package domain

import (
	"time"

	"gorm.io/gorm"
)

// Candidate represents a candidate for election.
type Candidate struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" swaggerignore:"true" json:"id"`
	Name           string         `gorm:"type:varchar(255);not null" example:"Beksultan" json:"name"`
	Photo          *string        `gorm:"type:varchar(255)" example:"link" json:"photo"`
	Education      *string        `gorm:"type:varchar(255)" example:"KBTU" json:"education"`
	Age            int            `gorm:"not null" example:"20" json:"age"`
	Party          *string        `gorm:"type:varchar(255)" example:"Jastar" json:"party"`
	Region         *string        `gorm:"type:varchar(255)" example:"SKO" json:"region"`
	Votes          int            `gorm:"default:0" swaggerignore:"true" json:"votes"`
	Type           CandidateType  `gorm:"type:varchar(255);not null" example:"manager" json:"type"`
	VotingStart    time.Time      `json:"voting_start" gorm:"type:datetime" example:"2025-11-12T09:00:00+05:00"`
	VotingDeadline time.Time      `json:"voting_deadline" gorm:"type:datetime" example:"2026-11-12T09:00:00+05:00"`
	DeletedAt      gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" swaggerignore:"true" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" swaggerignore:"true" json:"updated_at"`
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
