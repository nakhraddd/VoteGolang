package connect

import (
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v7"
)

func ConnectElasticsearch() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://elasticsearch:9200", // matches docker-compose service name
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Test connection
	res, err := es.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch connection failed: %s", res.String())
	}

	log.Println("Connected to Elasticsearch")
	return es, nil
}
