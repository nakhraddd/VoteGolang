package domain

import (
	"time"

	"gorm.io/gorm"
)

// Petition represents a created petition.
type Petition struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" swaggerignore:"true" json:"id"`
	UserID         uint           `gorm:"not null" swaggerignore:"true" json:"user_id"`
	Title          string         `gorm:"type:varchar(255);not null" example:"petition title" json:"title"`
	Photo          *string        `gorm:"type:varchar(255)" example:"link" json:"photo"`
	Description    *string        `gorm:"type:text" example:"petition description" json:"description"`
	VotesInFavor   int            `gorm:"default:0" swaggerignore:"true" json:"votes_in_favor"`
	VotesAgainst   int            `gorm:"default:0" swaggerignore:"true" json:"votes_against"`
	Goal           int            `gorm:"not null" example:"0" json:"goal"`
	VotingDeadline time.Time      `json:"voting_deadline" gorm:"type:datetime" example:"2025-05-10T23:59:00+05:00"`
	DeletedAt      gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" swaggerignore:"true" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" swaggerignore:"true" json:"updated_at"`
}

type PetitionRepository interface {
	Create(petition *Petition) error
	GetAll() ([]Petition, error)
	GetAllPaginated(limit, offset int) ([]Petition, error)
	GetByID(id uint) (*Petition, error)
	VoteInFavor(id uint) error
	VoteAgainst(id uint) error
	Delete(id uint) error
}

type PetitionVote struct {
	ID         uint           `gorm:"primaryKey;autoIncrement"`
	UserID     uint           `gorm:"not null;uniqueIndex:idx_user_petition" swaggerignore:"true"`
	PetitionID uint           `gorm:"not null;uniqueIndex:idx_user_petition"`
	VoteType   VoteType       `gorm:"type:varchar(255);not null"`
	DeletedAt  gorm.DeletedAt `json:"-" swaggerignore:"true"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
}

// PetitionVoteRequest represents the body to petition for a petition petition request.
type PetitionVoteRequest struct {
	UserId     uint `json:"user_id" swaggerignore:"true"`
	PetitionID uint `json:"petition_id"`
	// Enum values: favor, against
	VoteType VoteType `json:"vote_type" enum:"favor,against" example:"favor, against"`
}

type VoteType string

const (
	Favor   VoteType = "favor"
	Against VoteType = "against"
)

func IsValidVoteType(v string) bool {
	return v == string(Favor) || v == string(Against)
}
