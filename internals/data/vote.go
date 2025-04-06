package data

import "gorm.io/gorm"

type Vote struct {
	gorm.Model
	UserID        string `json:"user_id"`
	CandidateID   uint   `json:"candidate_id"`
	CandidateType string `json:"candidate_type"` // "president", "deputy", "session_deputy"
}
