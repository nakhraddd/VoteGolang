package candidate_data

// VoteRequest represents the body to vote_request for a candidate.
// swagger:model
type VoteRequest struct {
	CandidateID   uint   `json:"candidate_id"`
	CandidateType string `json:"candidate_type" enums:"presidential,deputy,manager" example:"presidential, deputy, manager"`
}
