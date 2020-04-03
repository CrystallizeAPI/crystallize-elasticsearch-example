package service

import (
	"context"

	"github.com/olivere/elastic/v7"
)

// SearchService holds all of the necessary methods for searching an index.
type SearchService struct{}

// Search performs the provided search query and returns the result.
func (s *SearchService) Search(ctx context.Context, client *elastic.Client, query *elastic.BoolQuery) ([]interface{}, error) {
	searchResult, err := client.Search().
		Index(CatalogueIndex).
		Query(query).
		Pretty(true).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	var items []interface{}
	for _, hit := range searchResult.Hits.Hits {
		items = append(items, hit.Source)
	}

	return items, nil
}
