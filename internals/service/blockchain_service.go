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
	FeeWei        int64       `json:"feeWei"`
}

// BlockchainService defines the interface for interacting with a blockchain.
type BlockchainService interface {
	LogCandidateCreation(candidate *domain.Candidate) (*TransactionLog, error)
	LogCandidateVote(userID uint, candidateID uint, candidateType domain.CandidateType) (*TransactionLog, error)
	LogPetitionCreation(petition *domain.Petition) (*TransactionLog, error)
	LogPetitionVote(userID uint, petitionID uint, voteType domain.VoteType) (*TransactionLog, error)

	GetServiceInfo() (map[string]interface{}, error)
}
