package candidate_usecase

import (
	"VoteGolang/internals/blockchain"
	candidate_data2 "VoteGolang/internals/domain"
	"errors"
	"time"
)

// CandidateUseCase handles business logic related to election candidates.
type CandidateUseCase struct {
	CandidateRepo candidate_data2.CandidateRepository
	VoteRepo      candidate_data2.VoteRepository
	Blockchain    *blockchain.Blockchain
}

func NewCandidateUseCase(cRepo candidate_data2.CandidateRepository, vRepo candidate_data2.VoteRepository, bc *blockchain.Blockchain) *CandidateUseCase {
	return &CandidateUseCase{
		CandidateRepo: cRepo,
		VoteRepo:      vRepo,
		Blockchain:    bc,
	}
}
func (uc *CandidateUseCase) CreateCandidate(candidate *candidate_data2.Candidate) error {
	if err := uc.CandidateRepo.Create(candidate); err != nil {
		return err
	}

	transaction := blockchain.Transaction{
		Type:    "CANDIDATE_CREATION",
		Payload: candidate,
	}
	uc.Blockchain.AddBlock(transaction)
	return nil
}

func (uc *CandidateUseCase) GetAllByTypePaginated(candidateType string, limit, offset int) ([]candidate_data2.Candidate, error) {
	return uc.CandidateRepo.GetAllByTypePaginated(candidateType, limit, offset)
}

// GetAllByType returns a list of candidates filtered by type.
func (uc *CandidateUseCase) GetAllByType(candidateType string) ([]candidate_data2.Candidate, error) {
	return uc.CandidateRepo.GetAllByType(candidateType)
}

// Vote votes for candidate by type, user_id, candidate_id.
func (uc *CandidateUseCase) Vote(candidateID uint, userID uint, candidateType candidate_data2.CandidateType) error {
	if !candidate_data2.IsValidCandidateType(string(candidateType)) {
		return errors.New("invalid candidate type")
	}

	voted, err := uc.VoteRepo.HasVoted(userID, string(candidateType))
	if err != nil {
		return err
	}
	if voted {
		return errors.New("already voted for this category")
	}

	candidate, err := uc.CandidateRepo.GetByID(candidateID)
	if err != nil {
		return err
	}

	if candidate.Type != candidateType {
		return errors.New("candidate type mismatch")
	}
	if time.Now().Before(candidate.VotingStart) {
		return errors.New("voting has not started for this candidate")
	}

	if time.Now().After(candidate.VotingDeadline) {
		return errors.New("voting period has ended for this candidate")
	}

	if err := uc.CandidateRepo.IncrementVote(candidateID); err != nil {
		return err
	}

	if err := uc.VoteRepo.SaveVote(candidateID, userID, string(candidateType)); err != nil {
		return err
	}

	transaction := blockchain.Transaction{
		Type: "VOTE_CAST",
		Payload: map[string]interface{}{
			"candidate_id":   candidateID,
			"user_id":        userID,
			"candidate_type": candidateType,
		},
	}
	uc.Blockchain.AddBlock(transaction)

	return nil
}
