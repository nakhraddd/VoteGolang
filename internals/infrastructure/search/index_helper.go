package search

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
)

// CreateIndexWithMapping ensures that an index exists with the correct mapping.
func CreateIndexWithMapping(es *elasticsearch.Client, indexName, mapping string) error {
	// Check if the index already exists
	res, err := es.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}
	// If the index exists (status 200), we're done.
	if res.StatusCode == http.StatusOK {
		return nil
	}
	// If the status is anything other than 404, it's an unexpected error.
	if res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code when checking for index: %d", res.StatusCode)
	}

	// Index does not exist, so create it with the mapping.
	res, err = es.Indices.Create(
		indexName,
		es.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	if res.IsError() {
		return fmt.Errorf("failed to create index with mapping: %s", res.String())
	}

	return nil
}
