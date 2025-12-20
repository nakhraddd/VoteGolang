package search

// Search interface defines the methods for searching.
type Search interface {
	Search(searchType, query string) ([]interface{}, error)
}
