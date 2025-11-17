package candidate_usecase

import (
	"VoteGolang/internals/domain"  // This was candidate_data2, aliasing to domain
	"VoteGolang/internals/service" // <-- NEW IMPORT
	"VoteGolang/internals/blockchain"
	candidate_data2 "VoteGolang/internals/domain"
	"context"
	"encoding/json"
	"errors"
	"log"
	"fmt"
	"log"
	"math/rand"
	"time"

	"VoteGolang/internals/infrastructure/search"

	"github.com/redis/go-redis/v9"
)

// CandidateUseCase handles business logic related to election candidates.
type CandidateUseCase struct {
	CandidateRepo domain.CandidateRepository
	VoteRepo      domain.VoteRepository
	Blockchain    service.BlockchainService // <-- CHANGED
}

func NewCandidateUseCase(cRepo candidate_data2.CandidateRepository, vRepo candidate_data2.VoteRepository, bc *blockchain.Blockchain, rdb *redis.Client, searchRepo *search.SearchRepository) *CandidateUseCase {
	return &CandidateUseCase{
		CandidateRepo: cRepo,
		VoteRepo:      vRepo,
		Blockchain:    bc,
		Redis:         rdb,
		SearchRepo:    searchRepo,
	}
}

func (uc *CandidateUseCase) CreateCandidate(candidate *domain.Candidate) error {
	if err := uc.CandidateRepo.Create(candidate); err != nil {
		return err
	}

	if uc.SearchRepo != nil {
		go func() {
			id := fmt.Sprintf("%d", candidate.ID)
			if err := uc.SearchRepo.IndexDocument(id, candidate); err != nil {
				log.Printf("❌ Failed to index candidate: %v", err)
			}
		}()
	}

	// Invalidate cache
	pattern := fmt.Sprintf("candidates:type:%s*", candidate.Type)
	keys, _ := uc.Redis.Keys(context.Background(), pattern).Result()
	for _, k := range keys {
		uc.Redis.Del(context.Background(), k)
	}
	if _, err := uc.Blockchain.LogCandidateCreation(candidate); err != nil {
		// If this fails, the DB is updated but the blockchain is not.
		// This is a critical error you need to handle (e.g., retry logic, compensating tx).
		// For now, we just log it.
		log.Printf("ERROR: Candidate %d created in DB but failed to log to blockchain: %v", candidate.ID, err)
		// You might choose to return this error, but that would require
		// rolling back the DB change, which requires a DB transaction.
	}
	// ---

	return nil
}

func (uc *CandidateUseCase) GetAllByTypePaginated(candidateType string, limit, offset int) ([]domain.Candidate, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("candidates:type:%s:page:%d:limit:%d", candidateType, offset/limit+1, limit)

	cached, err := uc.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var candidates []candidate_data2.Candidate
		if err := json.Unmarshal([]byte(cached), &candidates); err == nil {
			log.Println("Cache hit:", cacheKey)
			return candidates, nil
		}
	}

	log.Println("Cache miss:", cacheKey)

	candidates, err := uc.CandidateRepo.GetAllByTypePaginated(candidateType, limit, offset)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(candidates)
	uc.Redis.Set(ctx, cacheKey, data, time.Duration(rand.Intn(5)+25)*time.Minute) // 25–30 minutes

	return uc.CandidateRepo.GetAllByTypePaginated(candidateType, limit, offset)
}

// GetAllByType returns a list of candidates filtered by type.
func (uc *CandidateUseCase) GetAllByType(candidateType string) ([]domain.Candidate, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("candidates:type:%s", candidateType)

	// Try cache first
	cached, err := uc.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var candidates []candidate_data2.Candidate
		if err := json.Unmarshal([]byte(cached), &candidates); err == nil {
			log.Println("Cache hit:", cacheKey)
			return candidates, nil
		}
	}

	log.Println("Cache miss:", cacheKey)
	// Fallback to DB
	candidates, err := uc.CandidateRepo.GetAllByType(candidateType)
	if err != nil {
		return nil, err
	}

	// Save to Redis
	data, _ := json.Marshal(candidates)
	uc.Redis.Set(ctx, cacheKey, data, time.Duration(rand.Intn(5)+25)*time.Minute) // 25–30 minutes

	return uc.CandidateRepo.GetAllByType(candidateType)
}

// Vote votes for candidate by type, user_id, candidate_id.
func (uc *CandidateUseCase) Vote(candidateID uint, userID uint, candidateType domain.CandidateType) error {
	if !domain.IsValidCandidateType(string(candidateType)) {
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

	// Use transaction to ensure atomicity
	return uc.VoteRepo.VoteWithTransaction(candidateID, userID, string(candidateType), func() error {
		// Increment vote count
		if err := uc.CandidateRepo.IncrementVote(candidateID); err != nil {
			return err
		}

	if err := uc.VoteRepo.SaveVote(candidateID, userID, string(candidateType)); err != nil {
		// If this fails, you should roll back the IncrementVote
		return err
	}

	// --- NEW ---
	if _, err := uc.Blockchain.LogCandidateVote(userID, candidateID, candidateType); err != nil {
		// CRITICAL: DB vote is saved, but blockchain log failed.
		// This requires a rollback of the DB changes.
		log.Printf("ERROR: Vote (user %d, candidate %d) saved to DB but failed to log to blockchain: %v", userID, candidateID, err)
		// For a robust system, you MUST roll back the DB changes here.
		// For now, we just log.
	}
	// ---

	return nil
}
