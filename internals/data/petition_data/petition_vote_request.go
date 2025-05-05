package petition_data

// PetitionVoteRequest represents the body to petition_data for a petition_data petition_data request.
type PetitionVoteRequest struct {
	UserId     uint     `json:"user_id"`
	PetitionID uint     `json:"petition_id"`
	VoteType   VoteType `json:"vote_type"`
}
