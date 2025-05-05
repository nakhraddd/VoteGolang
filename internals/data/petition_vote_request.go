package data

// PetitionVoteRequest represents the body to vote for a petition vote request.
type PetitionVoteRequest struct {
	UserId     uint   `json:"user_id"`
	PetitionID uint   `json:"petition_id"`
	VoteType   string `json:"vote_type"` // TODO: types.VoteType
}
