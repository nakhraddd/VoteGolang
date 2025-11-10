package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v7"
)

type SearchRepository struct {
	Client *elasticsearch.Client
	Index  string
}

func NewSearchRepository(client *elasticsearch.Client, index string) *SearchRepository {
	return &SearchRepository{Client: client, Index: index}
}

// IndexDocument saves or updates a document in Elasticsearch
func (s *SearchRepository) IndexDocument(id string, data interface{}) error {
	body, _ := json.Marshal(data)
	res, err := s.Client.Index(
		s.Index,
		bytes.NewReader(body),
		s.Client.Index.WithDocumentID(id),
		s.Client.Index.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}
	log.Printf("✅ Document indexed [%s] in %s", id, s.Index)
	return nil
}

// Search performs full-text search on a specific field
func (s *SearchRepository) Search(query string, field string) ([]map[string]interface{}, error) {
	// Check if client is initialized
	if s == nil || s.Client == nil {
		return nil, fmt.Errorf("Search service unavailable")
	}

	// Build the search request
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				field: query,
			},
		},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	// Perform search
	res, err := s.Client.Search(
		s.Client.Search.WithContext(context.Background()),
		s.Client.Search.WithIndex(s.Index),
		s.Client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	// Parse hits
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	var items []map[string]interface{}
	for _, hit := range hits {
		item := hit.(map[string]interface{})["_source"].(map[string]interface{})
		items = append(items, item)
	}
	return items, nil
}
