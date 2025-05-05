package petition_data

type VoteType string

const (
	Favor   VoteType = "favor"
	Against VoteType = "against"
)

func IsValidVoteType(v string) bool {
	return v == string(Favor) || v == string(Against)
}
