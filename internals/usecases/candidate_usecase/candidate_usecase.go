package candidate_usecase

import (
	"VoteGolang/internals/app/logging"
	"VoteGolang/internals/domain"
	candidate_data2 "VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/repositories"
	"VoteGolang/internals/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

// CandidateUseCase handles business logic related to election candidates.
type CandidateUseCase struct {
	CandidateRepo domain.CandidateRepository
	VoteRepo      domain.VoteRepository
	Blockchain    service.BlockchainService
	Redis         *redis.Client
	SearchRepo    *repositories.SearchRepository
	Logger        *logging.KafkaLogger
}

func NewCandidateUseCase(
	cRepo candidate_data2.CandidateRepository,
	vRepo candidate_data2.VoteRepository,
	bc service.BlockchainService,
	rdb *redis.Client,
	searchRepo *repositories.SearchRepository,
	kafkaLogger *logging.KafkaLogger) *CandidateUseCase {
	return &CandidateUseCase{
		CandidateRepo: cRepo,
		VoteRepo:      vRepo,
		Blockchain:    bc,
		Redis:         rdb,
		SearchRepo:    searchRepo,
		Logger:        kafkaLogger,
	}
}

func (uc *CandidateUseCase) CreateCandidate(candidate *domain.Candidate) error {
	if err := uc.CandidateRepo.Create(candidate); err != nil {
		uc.Logger.Log("ERROR", fmt.Sprintf("Failed to create candidate in DB: %v", err))
		return err
	}
	uc.Logger.Log("INFO", fmt.Sprintf("Candidate %d created successfully in DB", candidate.ID))

	if uc.SearchRepo != nil {
		go func() {
			id := fmt.Sprintf("%d", candidate.ID)
			if err := uc.SearchRepo.Index(context.Background(), id, candidate); err != nil {
				uc.Logger.Log("WARN", fmt.Sprintf("Failed to index candidate %d: %v", candidate.ID, err))
			} else {
				uc.Logger.Log("DEBUG", fmt.Sprintf("Candidate %d indexed for search", candidate.ID))
			}
		}()
	}

	// Invalidate cache
	pattern := fmt.Sprintf("candidates:type:%s*", candidate.Type)
	keys, err := uc.Redis.Keys(context.Background(), pattern).Result()
	if err != nil {
		uc.Logger.Log("WARN", fmt.Sprintf("Failed to get keys for cache invalidation, pattern %s: %v", pattern, err))
	}
	for _, k := range keys {
		uc.Redis.Del(context.Background(), k)
	}
	uc.Logger.Log("DEBUG", fmt.Sprintf("Cache invalidated for pattern %s (keys: %d)", pattern, len(keys)))

	// Log to blockchain
	if _, err := uc.Blockchain.LogCandidateCreation(candidate); err != nil {

		uc.Logger.Log("ERROR", fmt.Sprintf("CRITICAL: Candidate %d created in DB but failed to log to blockchain: %v", candidate.ID, err))
	} else {
		uc.Logger.Log("INFO", fmt.Sprintf("Candidate %d logged to blockchain", candidate.ID))
	}

	return nil
}

func (uc *CandidateUseCase) GetAllByTypePaginated(candidateType string, limit, offset int) ([]domain.Candidate, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("candidates:type:%s:page:%d:limit:%d", candidateType, offset/limit+1, limit)

	cached, err := uc.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var candidates []domain.Candidate
		if err := json.Unmarshal([]byte(cached), &candidates); err == nil {
			uc.Logger.Log("DEBUG", fmt.Sprintf("Cache hit for: %s", cacheKey))
			return candidates, nil
		}
	}

	uc.Logger.Log("DEBUG", fmt.Sprintf("Cache miss for: %s", cacheKey))

	candidates, err := uc.CandidateRepo.GetAllByTypePaginated(candidateType, limit, offset)
	if err != nil {
		uc.Logger.Log("ERROR", fmt.Sprintf("Failed to get candidates from DB (type: %s): %v", candidateType, err))
		return nil, err
	}

	data, _ := json.Marshal(candidates)
	uc.Redis.Set(ctx, cacheKey, data, time.Duration(rand.Intn(5)+25)*time.Minute)

	return candidates, nil
}

// GetAllByType returns a list of candidates filtered by type.
func (uc *CandidateUseCase) GetAllByType(candidateType string) ([]domain.Candidate, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("candidates:type:%s", candidateType)

	// Try cache first
	cached, err := uc.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var candidates []domain.Candidate
		if err := json.Unmarshal([]byte(cached), &candidates); err == nil {
			uc.Logger.Log("DEBUG", fmt.Sprintf("Cache hit for: %s", cacheKey))
			return candidates, nil
		}
	}

	uc.Logger.Log("DEBUG", fmt.Sprintf("Cache miss for: %s", cacheKey))
	// Fallback to DB
	candidates, err := uc.CandidateRepo.GetAllByType(candidateType)
	if err != nil {
		uc.Logger.Log("ERROR", fmt.Sprintf("Failed to get candidates from DB (type: %s): %v", candidateType, err))
		return nil, err
	}

	// Save to Redis
	data, _ := json.Marshal(candidates)
	uc.Redis.Set(ctx, cacheKey, data, time.Duration(rand.Intn(5)+25)*time.Minute)

	return candidates, nil
}

// Vote votes for candidate by type, user_id, candidate_id.
func (uc *CandidateUseCase) Vote(candidateID uint, userID uint, candidateType domain.CandidateType) error {
	if !domain.IsValidCandidateType(string(candidateType)) {
		return errors.New("invalid candidate type")
	}

	voted, err := uc.VoteRepo.HasVoted(userID, string(candidateType))
	if err != nil {
		uc.Logger.Log("ERROR", fmt.Sprintf("Failed to check HasVoted (user: %d, type: %s): %v", userID, candidateType, err))
		return err
	}
	if voted {
		return errors.New("already voted for this category")
	}

	candidate, err := uc.CandidateRepo.GetByID(candidateID)
	if err != nil {
		uc.Logger.Log("WARN", fmt.Sprintf("Failed to find candidate %d for voting: %v", candidateID, err))
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

	//    This callback will be executed by VoteWithTransaction.
	dbTransactionCallback := func() error {
		// Increment vote count in the Candidates table
		if err := uc.CandidateRepo.IncrementVote(candidateID); err != nil {
			return err
		}
		// The VoteWithTransaction will handle the SaveVote part.
		return nil
	}

	//    This method should handle Begin, Commit, and Rollback.
	//    It runs the callback and then saves the vote in one atomic operation.
	err = uc.VoteRepo.VoteWithTransaction(candidateID, userID, string(candidateType), dbTransactionCallback)
	if err != nil {
		uc.Logger.Log("ERROR", fmt.Sprintf("DB transaction for vote failed (user: %d, candidate: %d): %v", userID, candidateID, err))
		return fmt.Errorf("database transaction failed: %w", err)
	}

	//    If this fails, the vote is *still valid* in our DB.
	if _, err := uc.Blockchain.LogCandidateVote(userID, candidateID, candidateType); err != nil {
		uc.Logger.Log("ERROR", fmt.Sprintf("CRITICAL: Vote (user %d, candidate %d) saved to DB but failed to log to blockchain: %v", userID, candidateID, err))
		// Do not return error, the vote was successful.
	} else {
		uc.Logger.Log("INFO", fmt.Sprintf("Vote (user %d, candidate %d) logged to blockchain", userID, candidateID))
	}

	return nil
}

func (uc *CandidateUseCase) GetCandidateByID(id uint) (*candidate_data2.Candidate, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("candidate:%d", id)

	if cached, err := uc.Redis.Get(ctx, cacheKey).Result(); err == nil {
		var candidate candidate_data2.Candidate
		if json.Unmarshal([]byte(cached), &candidate) == nil {
			log.Println("Cache hit:", cacheKey)
			return &candidate, nil
		}
	}

	log.Println("Cache miss:", cacheKey)
	candidate, err := uc.CandidateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(candidate)
	uc.Redis.Set(ctx, cacheKey, data, 5*time.Minute)
	return candidate, nil
}

func (uc *CandidateUseCase) DeleteCandidate(id uint) error {
	if err := uc.CandidateRepo.DeleteByID(id); err != nil {
		return err
	}

	// Invalidate all candidate caches
	ctx := context.Background()
	var cursor uint64
	for {
		keys, nextCursor, _ := uc.Redis.Scan(ctx, cursor, "candidates*", 100).Result()
		for _, k := range keys {
			uc.Redis.Del(ctx, k)
		}
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return nil
}
