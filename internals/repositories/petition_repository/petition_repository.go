package petition_repository

import (
	"VoteGolang/internals/data/petition_data"
	"gorm.io/gorm"
)

type PetitionRepository interface {
	Create(petition *petition_data.Petition) error
	GetAll() ([]petition_data.Petition, error)
	GetByID(id uint) (*petition_data.Petition, error)
	VoteInFavor(id uint) error
	VoteAgainst(id uint) error
	Delete(id uint) error
}

type petitionRepository struct {
	db *gorm.DB
}

func NewPetitionRepository(db *gorm.DB) PetitionRepository {
	return &petitionRepository{db: db}
}

func (r *petitionRepository) Create(petition *petition_data.Petition) error {
	return r.db.Create(petition).Error
}

func (r *petitionRepository) GetAll() ([]petition_data.Petition, error) {
	var petitions []petition_data.Petition
	err := r.db.Find(&petitions).Error
	return petitions, err
}

func (r *petitionRepository) GetByID(id uint) (*petition_data.Petition, error) {
	var petition petition_data.Petition
	err := r.db.First(&petition, id).Error
	if err != nil {
		return nil, err
	}
	return &petition, nil
}

func (r *petitionRepository) VoteInFavor(id uint) error {
	return r.db.Model(&petition_data.Petition{}).
		Where("id = ?", id).
		Update("favor", gorm.Expr("votes_in_favor + ?", 1)).
		Error
}

func (r *petitionRepository) VoteAgainst(id uint) error {
	return r.db.Model(&petition_data.Petition{}).
		Where("id = ?", id).
		Update("against", gorm.Expr("votes_against + ?", 1)).
		Error
}

func (r *petitionRepository) Delete(id uint) error {
	return r.db.Delete(&petition_data.Petition{}, id).Error
}
