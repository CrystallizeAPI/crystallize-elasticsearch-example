package service

import (
	"context"
	"fmt"
	"os"

	"github.com/crystallizeapi/crystallize-elasticsearch-example/types"
	"github.com/olivere/elastic/v7"
)

var (
	// CatalogueIndex is the elasticsearch index for the catalogue
	CatalogueIndex = "catalogue"

	// AttributesIndex is the elasticsearch index for variant attributes
	AttributesIndex = "attributes"
)

// IndexService holds all of the necesary methods for indexing the catalogue.
type IndexService struct{}

// CreateClient creates a new elastic client.
func CreateClient() (*elastic.Client, error) {
	url := os.Getenv("ELASTICSEARCH_NODE")
	user := os.Getenv("ELASTICSEARCH_USER")
	pass := os.Getenv("ELASTICSEARCH_PASS")

	return elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetBasicAuth(user, pass),
		elastic.SetSniff(false),
	)
}

// CreateIndex creates a new index with the specified name.s
func (i *IndexService) CreateIndex(ctx context.Context, client *elastic.Client, name string) error {
	res, err := client.CreateIndex(name).Do(ctx)
	if err != nil {
		return err
	}
	if !res.Acknowledged {
		return fmt.Errorf("Creating index failed: %s", name)
	}
	return nil
}

// DeleteIndex deletes an index with the specified name.
func (i *IndexService) DeleteIndex(ctx context.Context, client *elastic.Client, name string) error {
	res, err := client.DeleteIndex(name).Do(ctx)
	if err != nil {
		return err
	}
	if !res.Acknowledged {
		return fmt.Errorf("Deleting index failed: %s", CatalogueIndex)
	}

	return nil
}

// IndexExists checks to see whether an index with a specified name already
// exists within ElasticSearch.
func (i *IndexService) IndexExists(ctx context.Context, client *elastic.Client, name string) (bool, error) {
	exists, err := client.IndexExists(name).Do(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// Index indexes a single CatalogueItem.
func (i *IndexService) Index(ctx context.Context, client *elastic.Client, item types.CatalogueItem) error {
	_, err := client.Index().
		Index(CatalogueIndex).
		Id(item.ID).
		BodyJson(item).
		Refresh("wait_for").
		Do(ctx)

	return err
}

// BulkIndex indexes CatalogueItems in bulk.
func (i *IndexService) BulkIndex(ctx context.Context, client *elastic.Client, items []types.ElasticProduct) error {
	// Build bulk index request
	bulk := client.Bulk().Index(CatalogueIndex)
	for _, item := range items {
		bulk.Add(elastic.NewBulkIndexRequest().Id(item.Variant.ID).Doc(item))
	}

	// Execute the bulk operation
	bulkResponse, err := bulk.Do(ctx)
	if err != nil {
		return err
	}
	if bulkResponse.Errors {
		return fmt.Errorf("Bulk index failed %+v", bulkResponse.Failed())
	}

	return nil
}
