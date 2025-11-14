package petittion_usecase

import (
	"VoteGolang/internals/blockchain"
	petition_data2 "VoteGolang/internals/domain"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
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
	redis            *redis.Client
}

func NewPetitionUseCase(pr petition_data2.PetitionRepository, pvr petition_data2.PetitionVoteRepository, bc *blockchain.Blockchain, rdb *redis.Client) PetitionUseCase {
	return &petitionUseCase{
		petitionRepo:     pr,
		petitionVoteRepo: pvr,
		blockchain:       bc,
		redis:            rdb,
	}
}
func (uc *petitionUseCase) GetAllPetitionsPaginated(limit, offset int) ([]petition_data2.Petition, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("petitions:page:%d:limit:%d", offset/limit+1, limit)

	cached, err := uc.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var petitions []petition_data2.Petition
		if err := json.Unmarshal([]byte(cached), &petitions); err == nil {
			log.Println("Cache hit:", cacheKey)
			return petitions, nil
		}
	}

	log.Println("Cache miss:", cacheKey)

	petitions, err := uc.petitionRepo.GetAllPaginated(limit, offset)
	if err != nil {
		return nil, err
	}

	bytes, _ := json.Marshal(petitions)
	uc.redis.Set(ctx, cacheKey, bytes, 10*time.Minute)
	return petitions, nil
}

func (uc *petitionUseCase) CreatePetition(p *petition_data2.Petition) error {
	if err := uc.petitionRepo.Create(p); err != nil {
		return err
	}

	// Invalidate cache
	pattern := "petitions*"
	ctx := context.Background()
	var cursor uint64
	for {
		keys, nextCursor, _ := uc.redis.Scan(ctx, cursor, pattern, 100).Result()
		for _, k := range keys {
			uc.redis.Del(ctx, k)
		}
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	transaction := blockchain.Transaction{
		Type:    "PETITION_CREATION",
		Payload: p,
	}
	uc.blockchain.AddBlock(transaction)

	return nil
}

func (uc *petitionUseCase) GetAllPetitions() ([]petition_data2.Petition, error) {
	ctx := context.Background()
	cacheKey := "petitions"

	// Try cache first
	cached, err := uc.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var petitions []petition_data2.Petition
		if err := json.Unmarshal([]byte(cached), &petitions); err == nil {
			log.Println("Cache hit:", cacheKey)
			return petitions, nil
		}
	}

	log.Println("Cache miss:", cacheKey)
	// Fallback to DB
	petitions, err := uc.petitionRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// Save to Redis
	data, _ := json.Marshal(petitions)
	uc.redis.Set(ctx, cacheKey, data, time.Duration(rand.Intn(5)+25)*time.Minute) // 25–30 minutes
	return petitions, nil
}

func (uc *petitionUseCase) GetPetitionByID(id uint) (*petition_data2.Petition, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("petition:%d", id)

	if cached, err := uc.redis.Get(ctx, cacheKey).Result(); err == nil {
		var petition petition_data2.Petition
		if json.Unmarshal([]byte(cached), &petition) == nil {
			log.Println("Cache hit:", cacheKey)
			return &petition, nil
		}
	}

	log.Println("Cache miss:", cacheKey)
	petition, err := uc.petitionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(petition)
	uc.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	return petition, nil
}

func (uc *petitionUseCase) Vote(userID uint, petitionID uint, voteType petition_data2.VoteType) error {
	// Validate petition exists and meets criteria BEFORE transaction
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

	// Use transaction with row locking for idempotency
	return uc.petitionVoteRepo.VoteWithTransaction(userID, petitionID, voteType, func() error {
		// Update vote count based on vote type
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

		// Add to blockchain
		go func() {
			if uc.blockchain != nil {
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
			}
		}()
		//if uc.blockchain != nil {
		//	transaction := blockchain.Transaction{
		//		Type: "PETITION_VOTE",
		//		Payload: map[string]interface{}{
		//			"petition_id": petitionID,
		//			"user_id":     userID,
		//			"vote_type":   voteType,
		//		},
		//		Description: fmt.Sprintf("User %d voted on petition %d", userID, petitionID),
		//		Timestamp:   time.Now(),
		//	}
		//	uc.blockchain.AddBlock(transaction)
		//}

		// Invalidate cache
		ctx := context.Background()
		cacheKey := fmt.Sprintf("petition:%d", petitionID)
		uc.redis.Del(ctx, cacheKey)

		// Invalidate paginated cache
		var cursor uint64
		for {
			keys, nextCursor, _ := uc.redis.Scan(ctx, cursor, "petitions*", 100).Result()
			for _, k := range keys {
				uc.redis.Del(ctx, k)
			}
			if nextCursor == 0 {
				break
			}
			cursor = nextCursor
		}

		return nil
	})
}

func (uc *petitionUseCase) DeletePetition(id uint) error {
	if err := uc.petitionRepo.Delete(id); err != nil {
		return err
	}

	// Invalidate all petition caches
	ctx := context.Background()
	var cursor uint64
	for {
		keys, nextCursor, _ := uc.redis.Scan(ctx, cursor, "petitions*", 100).Result()
		for _, k := range keys {
			uc.redis.Del(ctx, k)
		}
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return nil
}

func (uc *petitionUseCase) HasUserVoted(userID uint, petitionID uint) (bool, error) {
	return uc.petitionVoteRepo.HasUserVoted(userID, petitionID)
}
