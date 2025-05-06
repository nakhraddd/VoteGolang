package petittion_usecase

import (
	"VoteGolang/internals/data/petition_data"
	"VoteGolang/internals/repositories/petition_repository"
	"VoteGolang/internals/repositories/votes_repositories"
	"fmt"
	"time"
)

// PetitionUseCase manages petition_data creation and retrieval.
type PetitionUseCase interface {
	// CreatePetition allows a user to create a new petition_data.
	CreatePetition(p *petition_data.Petition) error
	// GetAllPetitions returns all active petitions.
	GetAllPetitions() ([]petition_data.Petition, error)
	GetPetitionByID(id uint) (*petition_data.Petition, error)
	Vote(userID uint, petitionID uint, voteType petition_data.VoteType) error
	DeletePetition(id uint) error
	HasUserVoted(userID uint, petitionID uint) (bool, error)
	GetAllPetitionsPaginated(limit, offset int) ([]petition_data.Petition, error)
}

type petitionUseCase struct {
	petitionRepo     petition_repository.PetitionRepository
	petitionVoteRepo votes_repositories.PetitionVoteRepository
}

func NewPetitionUseCase(pr petition_repository.PetitionRepository, pvr votes_repositories.PetitionVoteRepository) PetitionUseCase {
	return &petitionUseCase{
		petitionRepo:     pr,
		petitionVoteRepo: pvr,
	}
}
func (uc *petitionUseCase) GetAllPetitionsPaginated(limit, offset int) ([]petition_data.Petition, error) {
	return uc.petitionRepo.GetAllPaginated(limit, offset)
}

func (uc *petitionUseCase) CreatePetition(p *petition_data.Petition) error {
	return uc.petitionRepo.Create(p)
}

func (uc *petitionUseCase) GetAllPetitions() ([]petition_data.Petition, error) {
	return uc.petitionRepo.GetAll()
}

func (uc *petitionUseCase) GetPetitionByID(id uint) (*petition_data.Petition, error) {
	return uc.petitionRepo.GetByID(id)
}

func (uc *petitionUseCase) Vote(userID uint, petitionID uint, voteType petition_data.VoteType) error {
	voted, err := uc.petitionVoteRepo.HasUserVoted(userID, petitionID)
	if err != nil {
		return err
	}
	if voted {
		return fmt.Errorf("user has already voted")
	}

	petition, err := uc.petitionRepo.GetByID(petitionID)
	if err != nil {
		return err
	}

	if !petition_data.IsValidVoteType(string(voteType)) {
		return fmt.Errorf("invalid petition_data type: must be 'favor' or 'against'")
	}

	if time.Now().After(petition.VotingDeadline) {
		return fmt.Errorf("voting period has ended")
	}

	totalVotes := petition.VotesInFavor + petition.VotesAgainst
	if totalVotes >= petition.Goal {
		return fmt.Errorf("petition_data goal has been reached")
	}

	vote := &petition_data.PetitionVote{
		UserID:     userID,
		PetitionID: petitionID,
		VoteType:   voteType,
	}

	err = uc.petitionVoteRepo.CreateVote(vote)
	if err != nil {
		return err
	}

	switch voteType {
	case petition_data.Favor:
		return uc.petitionRepo.VoteInFavor(petitionID)
	case petition_data.Against:
		return uc.petitionRepo.VoteAgainst(petitionID)
	default:
		return fmt.Errorf("invalid petition_data type")
	}

}

func (uc *petitionUseCase) DeletePetition(id uint) error {
	return uc.petitionRepo.Delete(id)
}

func (uc *petitionUseCase) HasUserVoted(userID uint, petitionID uint) (bool, error) {
	return uc.petitionVoteRepo.HasUserVoted(userID, petitionID)
}
