package service

import (
	"context"
	"reflect"

	"github.com/crystallizeapi/crystallize-elasticsearch-example/types"
	"github.com/olivere/elastic/v7"
)

type SearchService struct{}

func (s *SearchService) Search(ctx context.Context, client *elastic.Client, query *elastic.BoolQuery) ([]types.CatalogueItem, error) {
	searchResult, err := client.Search().
		Index(CatalogueIndex).
		Query(query).
		Pretty(true).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	var catalogueItems []types.CatalogueItem
	var ci types.CatalogueItem
	for _, item := range searchResult.Each(reflect.TypeOf(ci)) {
		if t, ok := item.(types.CatalogueItem); ok {
			catalogueItems = append(catalogueItems, t)
		}
	}

	return catalogueItems, nil
}
