package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tidwall/gjson"
)

// Elasticsearch is an implementation of the Search interface that uses Elasticsearch.
type Elasticsearch struct {
	Address string
}

// NewElasticsearch creates a new Elasticsearch instance.
func NewElasticsearch(address string) *Elasticsearch {
	return &Elasticsearch{Address: address}
}

// SearchCandidates searches for candidates by name.
func (e *Elasticsearch) SearchCandidates(query string) ([]interface{}, error) {
	return e.search("candidates", "name", query)
}

// SearchPetitions searches for petitions by title.
func (e *Elasticsearch) SearchPetitions(query string) ([]interface{}, error) {
	return e.search("petitions", "title", query)
}

func (e *Elasticsearch) search(index, field, query string) ([]interface{}, error) {
	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"fuzzy": map[string]interface{}{
				field: map[string]interface{}{
					"value":     query,
					"fuzziness": "AUTO",
				},
			},
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/%s/_search", e.Address, index), "application/json", bytes.NewBuffer(reqJSON))
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
