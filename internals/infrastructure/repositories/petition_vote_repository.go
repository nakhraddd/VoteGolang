package repositories

import (
	petition_data2 "VoteGolang/internals/domain"

	"gorm.io/gorm"
)

type petitionVoteGormRepository struct {
	db *gorm.DB
}

func NewPetitionVoteRepository(db *gorm.DB) petition_data2.PetitionVoteRepository {
	return &petitionVoteGormRepository{db: db}
}

func (r *petitionVoteGormRepository) CreateVote(vote *petition_data2.PetitionVote) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(vote).Error; err != nil {
			return err
		}

		var column string
		switch vote.VoteType {
		case petition_data2.Favor:
			column = "votes_in_favor"
		case petition_data2.Against:
			column = "votes_against"
		default:
			return nil
		}

		return tx.Model(&petition_data2.Petition{}).
			Where("id = ?", vote.PetitionID).
			UpdateColumn(column, gorm.Expr(column+" + ?", 1)).
			Error
	})
}

func (r *petitionVoteGormRepository) HasUserVoted(userID uint, petitionID uint) (bool, error) {
	var count int64
	err := r.db.Model(&petition_data2.PetitionVote{}).
		Where("user_id = ? AND petition_id = ?", userID, petitionID).
		Count(&count).Error

	return count > 0, err
}
