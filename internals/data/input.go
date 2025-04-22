package data

// Input represents the body to vote for a petition vote request.
var Input struct {
	UserId     uint   `json:"user_id"`
	PetitionID uint   `json:"petition_id"`
	VoteType   string `json:"vote_type"`
}
