package tasks

import (
	"context"
	"fmt"

	"github.com/crystallizeapi/crystallize-elasticsearch-example/service"
	"github.com/crystallizeapi/crystallize-elasticsearch-example/types"
	"github.com/machinebox/graphql"
	"github.com/olivere/elastic/v7"
)

type CatalogueBulkIndexTask struct {
	catalogueItems []types.CatalogueItem
	client         *elastic.Client
	tenant         string
}

// NewCatalogueBulkIndexTask creates a new CatalogueBulkIndexTask.
func NewCatalogueBulkIndexTask(tenant string) (*CatalogueBulkIndexTask, error) {
	if tenant == "" {
		return nil, fmt.Errorf("You must provide a tenant identifier to index")
	}

	client, err := service.CreateClient()
	if err != nil {
		return nil, err
	}

	return &CatalogueBulkIndexTask{
		catalogueItems: []types.CatalogueItem{},
		client:         client,
		tenant:         tenant,
	}, nil
}

// CatalogueQuery is the query to fetch all the items in a catalogue relating to
// a tenant.
var CatalogueQuery = `
	query {
		catalogue(path: "/", language: "en") {
			children {
				...item
				children {
					...item
					children {
						...item
					}
				}
			}
		}
	}

	fragment item on Item {
		id
		name
		type
	}
`

// CatalogueResponse represents the GraphQL response of the catalogue query.
// In this example it is nested to the second layer of children, but it can be
// nested as much as necessary based on the CatalogueQuery.
type CatalogueResponse struct {
	Catalogue struct {
		Children []struct {
			types.CatalogueItem
			Children []struct {
				types.CatalogueItem
				Children []struct {
					types.CatalogueItem
				}
			}
		}
	}
}

// normaliseCatalogueItems normalises the response data into a flat array of items.
func normaliseCatalogueItems(respData CatalogueResponse) []types.CatalogueItem {
	var catalogueItems []types.CatalogueItem

	for _, item := range respData.Catalogue.Children {
		catalogueItem := types.CatalogueItem{
			ID:   item.ID,
			Name: item.Name,
			Type: item.Type,
		}
		catalogueItems = append(catalogueItems, catalogueItem)

		for _, item2 := range item.Children {
			catalogueItem = types.CatalogueItem{
				ID:   item2.ID,
				Name: item2.Name,
				Type: item2.Type,
			}
			catalogueItems = append(catalogueItems, catalogueItem)

			for _, item3 := range item2.Children {
				catalogueItem = types.CatalogueItem{
					ID:   item3.ID,
					Name: item3.Name,
					Type: item3.Type,
				}
				catalogueItems = append(catalogueItems, catalogueItem)
			}
		}
	}

	return catalogueItems
}

// Setup fetches the catalogue items to be indexed from Crystallize's catalogue
// API via GraphQL.
func (t *CatalogueBulkIndexTask) Setup(ctx context.Context) error {
	apiUrl := fmt.Sprintf("https://api.crystallize.com/%s/catalogue", t.tenant)
	graphqlClient := graphql.NewClient(apiUrl)
	req := graphql.NewRequest(CatalogueQuery)

	// Query the catalogue API
	var respData CatalogueResponse
	if err := graphqlClient.Run(ctx, req, &respData); err != nil {
		return err
	}

	// Normalise catalogue
	catalogueItems := normaliseCatalogueItems(respData)
	t.catalogueItems = catalogueItems

	fmt.Printf("Queried %d items\n", len(t.catalogueItems))

	return nil
}

// Execute re-creates the catalogue index and indexes the catalogue items.
func (t *CatalogueBulkIndexTask) Execute(ctx context.Context) error {
	indexService := service.IndexService{}

	// Check whether the catalogue index exists
	exists, err := indexService.IndexExists(ctx, t.client, service.CatalogueIndex)
	if err != nil {
		return err
	}

	// Delete the catalogue index if it exists
	if exists {
		if err = indexService.DeleteIndex(ctx, t.client, service.CatalogueIndex); err != nil {
			return err
		}
	}

	// Create the catsalogue index
	if err = indexService.CreateIndex(ctx, t.client, service.CatalogueIndex); err != nil {
		return err
	}

	fmt.Printf("Indexing %d items\n", len(t.catalogueItems))

	// Bulk index the catalogue items
	return indexService.BulkIndex(ctx, t.client, t.catalogueItems)
}
