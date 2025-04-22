package data

// VoteRequest represents the body to vote for a candidate.
type VoteRequest struct {
	CandidateID   uint   `json:"candidate_id"`
	CandidateType string `json:"candidate_type"`
}
