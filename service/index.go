package service

import (
	"context"
	"fmt"

	"github.com/crystallizeapi/crystallize-elasticsearch-example/types"
	"github.com/olivere/elastic/v7"
)

var (
	CatalogueIndex = "catalogue"
)

type IndexService struct{}

// CreateClient creates a new elastic client.
func CreateClient() (*elastic.Client, error) {
	return elastic.NewClient(elastic.SetURL("http://localhost:9200"), elastic.SetSniff(false))
}

// CreateIndex creates a new index with the specified name.s
func (i *IndexService) CreateIndex(ctx context.Context, client *elastic.Client, name string) error {
	res, err := client.CreateIndex(name).Do(ctx)
	if err != nil {
		return err
	}
	if !res.Acknowledged {
		return fmt.Errorf("Creating index failed: %s\n", name)
	}
	return nil
}

// DeleteIndex deletes an index with the specified name.
func (i *IndexService) DeleteIndex(ctx context.Context, client *elastic.Client, name string) error {
	res, err := client.DeleteIndex(CatalogueIndex).Do(ctx)
	if err != nil {
		return err
	}
	if !res.Acknowledged {
		return fmt.Errorf("Deleting index failed: %s\n", CatalogueIndex)
	}

	return nil
}

// IndexExists checks to see whether an index with a specified name already
// exists within ElasticSearch.
func (i *IndexService) IndexExists(ctx context.Context, client *elastic.Client, name string) (bool, error) {
	exists, err := client.IndexExists(CatalogueIndex).Do(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// BulkIndex indexes CatalogueItems in bulk.
func (i *IndexService) BulkIndex(ctx context.Context, client *elastic.Client, items []types.CatalogueItem) error {
	// Build bulk index request
	bulk := client.Bulk().Index(CatalogueIndex)
	for _, item := range items {
		bulk.Add(elastic.NewBulkIndexRequest().Id(item.ID).Doc(item))
	}

	// Execute the bulk operation
	bulkResponse, err := bulk.Do(ctx)
	if err != nil {
		return err
	}
	if bulkResponse.Errors {
		return fmt.Errorf("Bulk index failed\n")
	}

	return nil
}
