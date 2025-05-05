package votes_repositories

import (
	"VoteGolang/internals/data/petition_data"
	"gorm.io/gorm"
)

// PetitionVoteRepository manages voting data for petitions.
type PetitionVoteRepository interface {
	CreateVote(vote *petition_data.PetitionVote) error
	HasUserVoted(userID uint, petitionID uint) (bool, error)
}

type petitionVoteRepository struct {
	db *gorm.DB
}

func NewPetitionVoteRepository(db *gorm.DB) PetitionVoteRepository {
	return &petitionVoteRepository{db: db}
}

func (r *petitionVoteRepository) CreateVote(vote *petition_data.PetitionVote) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(vote).Error; err != nil {
			return err
		}

		var column string
		switch vote.VoteType {
		case petition_data.Favor:
			column = "votes_in_favor"
		case petition_data.Against:
			column = "votes_against"
		default:
			return nil
		}

		return tx.Model(&petition_data.Petition{}).
			Where("id = ?", vote.PetitionID).
			UpdateColumn(column, gorm.Expr(column+" + ?", 1)).
			Error
	})
}

func (r *petitionVoteRepository) HasUserVoted(userID uint, petitionID uint) (bool, error) {
	var count int64
	err := r.db.Model(&petition_data.PetitionVote{}).
		Where("user_id = ? AND petition_id = ?", userID, petitionID).
		Count(&count).Error

	return count > 0, err
}
