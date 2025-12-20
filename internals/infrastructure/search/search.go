package search

// Search interface defines the methods for searching candidates and petitions.
type Search interface {
	SearchCandidates(query string) ([]interface{}, error)
	SearchPetitions(query string) ([]interface{}, error)
}
