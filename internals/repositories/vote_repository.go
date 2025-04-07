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
	exists, err := r.HasUserVotedForCandidate(vote.UserID, vote.CandidateID, vote.CandidateType)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("user has already voted for this candidate")
	}

	if err := r.db.Create(vote).Error; err != nil {
		return err
	}

	return r.db.Model(&data.Candidate{}).
		Where("id = ? AND type = ?", vote.CandidateID, vote.CandidateType).
		Update("votes", gorm.Expr("votes + ?", 1)).Error
}

func (r *voteRepository) HasUserVotedForCandidate(userID string, candidateID uint, candidateType string) (bool, error) {
	var count int64
	err := r.db.Model(&data.Vote{}).
		Where("user_id = ? AND candidate_id = ? AND candidate_type = ?", userID, candidateID, candidateType).
		Count(&count).Error
	return count > 0, err
}
