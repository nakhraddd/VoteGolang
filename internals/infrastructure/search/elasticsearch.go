package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tidwall/gjson"
)

type searchConfig struct {
	Index string
	Field string
}

// Elasticsearch is an implementation of the Search interface that uses Elasticsearch.
type Elasticsearch struct {
	Address          string
	searchTypeConfig map[string]searchConfig
}

// NewElasticsearch creates a new Elasticsearch instance.
func NewElasticsearch(address string) *Elasticsearch {
	return &Elasticsearch{
		Address: address,
		searchTypeConfig: map[string]searchConfig{
			"candidates": {Index: "candidates", Field: "name"},
			"petitions":  {Index: "petitions", Field: "title"},
		},
	}
}

// Search performs a search on the specified index and field.
// If the query is empty, it returns all documents.
func (e *Elasticsearch) Search(searchType, query string) ([]interface{}, error) {
	config, ok := e.searchTypeConfig[searchType]
	if !ok {
		return nil, fmt.Errorf("unknown search type: %s", searchType)
	}

	var reqBody map[string]interface{}
	if query == "" {
		// If the query is empty, use match_all to return all documents.
		reqBody = map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
			"size": 100, // Add a reasonable size limit to avoid overwhelming the client.
		}
	} else {
		// Otherwise, use the match query for searching.
		reqBody = map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					config.Field: query,
				},
			},
		}
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/%s/_search", e.Address, config.Index), "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Elasticsearch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Elasticsearch request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	hits := gjson.GetBytes(body, "hits.hits").Array()
	results := make([]interface{}, len(hits))
	for i, hit := range hits {
		results[i] = hit.Get("_source").Value()
	}

	return results, nil
}
