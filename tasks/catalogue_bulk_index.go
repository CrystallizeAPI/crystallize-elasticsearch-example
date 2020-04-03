package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CrystallizeAPI/crystallize-elasticsearch-example/service"
	esTypes "github.com/CrystallizeAPI/crystallize-elasticsearch-example/types"
	"github.com/CrystallizeAPI/crystallize-go-types/types"
	"github.com/machinebox/graphql"
	"github.com/olivere/elastic/v7"
)

type CatalogueBulkIndexTask struct {
	items  []esTypes.ElasticProduct
	client *elastic.Client
	tenant string
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
		items:  []esTypes.ElasticProduct{},
		client: client,
		tenant: tenant,
	}, nil
}

// CatalogueQuery is the query to fetch all the items in a catalogue relating to
// a tenant.
var CatalogueQuery = `
	query {
		catalogue(path: "/", language: "en") {
			children {
				...item
				...product
				children {
					...item
					...product
					children {
						...item
						...product
						children {
							...item
							...product
							children {
							...item
							...product
						}
						}
					}
				}
			}
		}
	}

	fragment item on Item {
		id
		name
		path
		type
		topics {
			id
			name
			parentId
		}
	}

	fragment product on Product {
		variants {
			id
			name
			sku
			price
			stock
			isDefault
			attributes {
				attribute
				value
			}
			images {
				key
				url
				variants {
					key
					url
					width
				}
			}
		}
	}
`

func getImageVariants(image types.Image) []types.ImageVariant {
	imageVariants := []types.ImageVariant{}

	for _, variant := range image.Variants {
		if variant.Width == 200 || variant.Width == 500 {
			imageVariants = append(imageVariants, variant)
		}
	}

	return imageVariants
}

func getImages(variant types.ProductVariant) *[]types.Image {
	if variant.Images != nil && len(*variant.Images) > 0 {
		images := *variant.Images
		imageVariants := getImageVariants(images[0])
		return &[]types.Image{types.Image{
			URL:      images[0].URL,
			Key:      images[0].Key,
			Variants: imageVariants,
		}}
	}
	return nil
}

func normaliseChildren(children []interface{}) ([]esTypes.ElasticProduct, error) {
	var items []esTypes.ElasticProduct

	for _, item := range children {
		catalogueItem := item.(map[string]interface{})
		jsonBody, err := json.Marshal(catalogueItem)
		if err != nil {
			return nil, err
		}

		if catalogueItem["type"] == "product" {
			product := types.Product{}
			if err := json.Unmarshal(jsonBody, &product); err != nil {
				return nil, err
			}
			for _, variant := range product.Variants {
				productVariants := []types.ProductVariant{}
				for _, v := range product.Variants {
					images := getImages(v)
					v.Images = images
					productVariants = append(productVariants, v)
				}

				images := getImages(variant)
				variant.Images = images
				product.Variants = productVariants

				elasticProduct := esTypes.ElasticProduct{
					Variant: variant,
					Product: product,
				}
				items = append(items, elasticProduct)
			}
		}

		if catalogueItem["children"] != nil {
			childItems, err := normaliseChildren(catalogueItem["children"].([]interface{}))
			if err != nil {
				return nil, err
			}

			items = append(items, childItems...)
		}
	}

	return items, nil
}

// Setup fetches the catalogue items to be indexed from Crystallize's catalogue
// API via GraphQL.
func (t *CatalogueBulkIndexTask) Setup(ctx context.Context) error {
	apiURL := fmt.Sprintf("https://api.crystallize.com/%s/catalogue", t.tenant)
	graphqlClient := graphql.NewClient(apiURL)
	req := graphql.NewRequest(CatalogueQuery)

	// Query the catalogue API
	var respData map[string]interface{}
	if err := graphqlClient.Run(ctx, req, &respData); err != nil {
		return err
	}

	catalogue := respData["catalogue"].(map[string]interface{})
	children := catalogue["children"].([]interface{})

	// Normalise catalogue
	items, err := normaliseChildren(children)
	if err != nil {
		return err
	}

	t.items = items
	fmt.Printf("Queried %d items\n", len(t.items))

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

	fmt.Printf("Indexing %d items\n", len(t.items))

	// Bulk index the catalogue items
	return indexService.BulkIndex(ctx, t.client, t.items)
}
