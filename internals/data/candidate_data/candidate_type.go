package candidate_data

type CandidateType string

const (
	Presidential CandidateType = "presidential"
	Deputy       CandidateType = "deputy"
	Manager      CandidateType = "manager"
)

func IsValidCandidateType(t string) bool {
	switch CandidateType(t) {
	case Presidential, Deputy, Manager:
		return true
	default:
		return false
	}
}
