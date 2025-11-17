package service

import (
	"VoteGolang/internals/domain"
	"time"
)

// TransactionLog represents a generic log of an on-chain action.
type TransactionLog struct {
	TransactionID string      `json:"transactionId"`
	Timestamp     time.Time   `json:"timestamp"`
	ActionType    string      `json:"actionType"`
	Details       interface{} `json:"details"`
	FeeWei        int64       `json:"feeWei"` // <-- RENAMED from FeeSUN
}

// BlockchainService defines the interface for interacting with a blockchain.
// This decouples our app from a specific implementation (local vs. TRON vs. BNB).
type BlockchainService interface {
	LogCandidateCreation(candidate *domain.Candidate) (*TransactionLog, error)
	LogCandidateVote(userID uint, candidateID uint, candidateType domain.CandidateType) (*TransactionLog, error)
	LogPetitionCreation(petition *domain.Petition) (*TransactionLog, error)
	LogPetitionVote(userID uint, petitionID uint, voteType domain.VoteType) (*TransactionLog, error)

	// This replaces GetChain(). It returns status info about the connection.
	GetServiceInfo() (map[string]interface{}, error)
}
