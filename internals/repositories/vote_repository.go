package repositories

import (
	"VoteGolang/internals/data"
	"errors"

	"gorm.io/gorm"
)

type VoteRepository interface {
	VoteForCandidate(vote *data.Vote) error
	HasUserVotedForCandidate(userID string, candidateID uint, candidateType string) (bool, error)
}

type voteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) VoteRepository {
	return &voteRepository{db: db}
}

func (r *voteRepository) VoteForCandidate(vote *data.Vote) error {
	// Check if user has already voted for this candidate
	exists, err := r.HasUserVotedForCandidate(vote.UserID, vote.CandidateID, vote.CandidateType)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("user has already voted for this candidate")
	}

	// Add vote record
	if err := r.db.Create(vote).Error; err != nil {
		return err
	}

	// Increment vote count in corresponding candidate table
	switch vote.CandidateType {
	case "president":
		return r.db.Model(&domain.President{}).Where("id = ?", vote.CandidateID).Update("votes", gorm.Expr("votes + ?", 1)).Error
	case "deputy":
		return r.db.Model(&domain.Deputy{}).Where("id = ?", vote.CandidateID).Update("votes", gorm.Expr("votes + ?", 1)).Error
	case "session_deputy":
		return r.db.Model(&domain.SessionDeputy{}).Where("id = ?", vote.CandidateID).Update("votes", gorm.Expr("votes + ?", 1)).Error
	default:
		return errors.New("invalid candidate type")
	}
}

func (r *voteRepository) HasUserVotedForCandidate(userID string, candidateID uint, candidateType string) (bool, error) {
	var count int64
	err := r.db.Model(&data.Vote{}).Where("user_id = ? AND candidate_id = ? AND candidate_type = ?", userID, candidateID, candidateType).Count(&count).Error
	return count > 0, err
}
