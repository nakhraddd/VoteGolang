package petittion_usecase

import (
	"VoteGolang/internals/blockchain"
	petition_data2 "VoteGolang/internals/domain"
	"fmt"
	"time"
)

// PetitionUseCase manages petition creation and retrieval.
type PetitionUseCase interface {
	// CreatePetition allows a user to create a new petition.
	CreatePetition(p *petition_data2.Petition) error
	// GetAllPetitions returns all active petitions.
	GetAllPetitions() ([]petition_data2.Petition, error)
	GetPetitionByID(id uint) (*petition_data2.Petition, error)
	Vote(userID uint, petitionID uint, voteType petition_data2.VoteType) error
	DeletePetition(id uint) error
	HasUserVoted(userID uint, petitionID uint) (bool, error)
	GetAllPetitionsPaginated(limit, offset int) ([]petition_data2.Petition, error)
}

type petitionUseCase struct {
	petitionRepo     petition_data2.PetitionRepository
	petitionVoteRepo petition_data2.PetitionVoteRepository
	blockchain       *blockchain.Blockchain
}

func NewPetitionUseCase(pr petition_data2.PetitionRepository, pvr petition_data2.PetitionVoteRepository, bc *blockchain.Blockchain) PetitionUseCase {
	return &petitionUseCase{
		petitionRepo:     pr,
		petitionVoteRepo: pvr,
		blockchain:       bc,
	}
}
func (uc *petitionUseCase) GetAllPetitionsPaginated(limit, offset int) ([]petition_data2.Petition, error) {
	return uc.petitionRepo.GetAllPaginated(limit, offset)
}

func (uc *petitionUseCase) CreatePetition(p *petition_data2.Petition) error {
	if err := uc.petitionRepo.Create(p); err != nil {
		return err
	}

	transaction := blockchain.Transaction{
		Type:    "PETITION_CREATION",
		Payload: p,
	}
	uc.blockchain.AddBlock(transaction)

	return nil
}

func (uc *petitionUseCase) GetAllPetitions() ([]petition_data2.Petition, error) {
	return uc.petitionRepo.GetAll()
}

func (uc *petitionUseCase) GetPetitionByID(id uint) (*petition_data2.Petition, error) {
	return uc.petitionRepo.GetByID(id)
}

func (uc *petitionUseCase) Vote(userID uint, petitionID uint, voteType petition_data2.VoteType) error {
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

	if !petition_data2.IsValidVoteType(string(voteType)) {
		return fmt.Errorf("invalid petition type: must be 'favor' or 'against'")
	}

	if time.Now().After(petition.VotingDeadline) {
		return fmt.Errorf("voting period has ended")
	}

	totalVotes := petition.VotesInFavor + petition.VotesAgainst
	if totalVotes >= petition.Goal {
		return fmt.Errorf("petition goal has been reached")
	}

	vote := &petition_data2.PetitionVote{
		UserID:     userID,
		PetitionID: petitionID,
		VoteType:   voteType,
	}

	err = uc.petitionVoteRepo.CreateVote(vote)
	if err != nil {
		return err
	}

	var dbErr error
	switch voteType {
	case petition_data2.Favor:
		dbErr = uc.petitionRepo.VoteInFavor(petitionID)
	case petition_data2.Against:
		dbErr = uc.petitionRepo.VoteAgainst(petitionID)
	default:
		return fmt.Errorf("invalid vote type")
	}

	if dbErr != nil {
		return dbErr
	}

	transaction := blockchain.Transaction{
		Type: "PETITION_VOTE",
		Payload: map[string]interface{}{
			"petition_id": petitionID,
			"user_id":     userID,
			"vote_type":   voteType,
		},
		Description: fmt.Sprintf("User %d voted on petition %d", userID, petitionID),
		Timestamp:   time.Now(),
	}
	uc.blockchain.AddBlock(transaction)

	return nil
}

func (uc *petitionUseCase) DeletePetition(id uint) error {
	return uc.petitionRepo.Delete(id)
}

func (uc *petitionUseCase) HasUserVoted(userID uint, petitionID uint) (bool, error) {
	return uc.petitionVoteRepo.HasUserVoted(userID, petitionID)
}
