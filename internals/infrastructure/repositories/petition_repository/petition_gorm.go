package petition_repository

import (
	"VoteGolang/internals/domain"

	"gorm.io/gorm"
)

type petitionGormRepository struct {
	db *gorm.DB
}

func NewPetitionRepository(db *gorm.DB) domain.PetitionRepository {
	return &petitionGormRepository{db: db}
}

func (r *petitionGormRepository) Create(petition *domain.Petition) error {
	return r.db.Create(petition).Error
}

func (r *petitionGormRepository) GetAllPaginated(limit, offset int) ([]domain.Petition, error) {
	var petitions []domain.Petition
	err := r.db.
		Limit(limit).
		Offset(offset).
		Find(&petitions).Error
	return petitions, err
}

func (r *petitionGormRepository) GetAll() ([]domain.Petition, error) {
	var petitions []domain.Petition
	err := r.db.Find(&petitions).Error
	return petitions, err
}

func (r *petitionGormRepository) GetByID(id uint) (*domain.Petition, error) {
	var petition domain.Petition
	err := r.db.First(&petition, id).Error
	if err != nil {
		return nil, err
	}
	return &petition, nil
}

func (r *petitionGormRepository) VoteInFavor(id uint) error {
	return r.db.Model(&domain.Petition{}).
		Where("id = ?", id).
		Update("votes_in_favor", gorm.Expr("votes_in_favor + ?", 1)).
		Error
}

func (r *petitionGormRepository) VoteAgainst(id uint) error {
	return r.db.Model(&domain.Petition{}).
		Where("id = ?", id).
		Update("votes_against", gorm.Expr("votes_against + ?", 1)).
		Error
}

func (r *petitionGormRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Petition{}, id).Error
}
