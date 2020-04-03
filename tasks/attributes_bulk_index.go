package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/crystallizeapi/crystallize-elasticsearch-example/service"
	"github.com/crystallizeapi/crystallize-elasticsearch-example/types"
	"github.com/machinebox/graphql"
	"github.com/olivere/elastic/v7"
)

type AttributesBulkIndexTask struct {
	attributes []VariantAttribute
	client     *elastic.Client
	tenant     string
}

func NewAttributesBulkIndexTask(tenant string) (*AttributesBulkIndexTask, error) {
	if tenant == "" {
		return nil, fmt.Errorf("You must provide a tenant identifier to index")
	}

	client, err := service.CreateClient()
	if err != nil {
		return nil, err
	}

	return &AttributesBulkIndexTask{
		attributes: []VariantAttribute{},
		client:     client,
		tenant:     tenant,
	}, nil
}

var AttributesQuery = `
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
	type
}

fragment product on Product {
	variants {
		attributes {
			attribute
			value
		}
	}
}
`

type VariantAttribute struct {
	Attribute string   `json:"attribute"`
	Values    []string `json:"values"`
}

func normaliseAttributes(children []interface{}, attributes []VariantAttribute) ([]VariantAttribute, error) {
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
				for _, variantAttr := range variant.Attributes {
					foundAttr := false
					for key, attr := range attributes {
						if attr.Attribute == variantAttr.Attribute {
							foundAttr = true
							foundVal := false
							for _, val := range attr.Values {
								if val == variantAttr.Value {
									foundVal = true
								}
							}

							if !foundVal {
								attributes[key].Values = append(attr.Values, variantAttr.Value)
							}
						}

					}
					if !foundAttr {
						attributes = append(attributes, VariantAttribute{Attribute: variantAttr.Attribute, Values: []string{variantAttr.Value}})
					}
				}
			}
		}

		if catalogueItem["children"] != nil {
			attributes, err = normaliseAttributes(catalogueItem["children"].([]interface{}), attributes)
			if err != nil {
				return nil, err
			}
		}
	}

	return attributes, nil
}

// Setup fetches the catalogue items to be indexed from Crystallize's catalogue
// API via GraphQL.
func (t *AttributesBulkIndexTask) Setup(ctx context.Context) error {
	apiURL := fmt.Sprintf("https://api.crystallize.com/%s/catalogue", t.tenant)
	graphqlClient := graphql.NewClient(apiURL)
	req := graphql.NewRequest(AttributesQuery)

	// Query the catalogue API
	var respData map[string]interface{}
	if err := graphqlClient.Run(ctx, req, &respData); err != nil {
		return err
	}

	catalogue := respData["catalogue"].(map[string]interface{})
	children := catalogue["children"].([]interface{})

	// Normalise catalogue
	attributes := []VariantAttribute{}
	attributes, err := normaliseAttributes(children, attributes)
	if err != nil {
		return err
	}

	t.attributes = attributes
	fmt.Printf("Queried %d attributes\n", len(t.attributes))

	return nil
}

// Execute re-creates the catalogue index and indexes the catalogue items.
func (t *AttributesBulkIndexTask) Execute(ctx context.Context) error {
	indexService := service.IndexService{}

	// Check whether the catalogue index exists
	exists, err := indexService.IndexExists(ctx, t.client, service.AttributesIndex)
	if err != nil {
		return err
	}

	// Delete the catalogue index if it exists
	if exists {
		if err = indexService.DeleteIndex(ctx, t.client, service.AttributesIndex); err != nil {
			return err
		}
	}

	// Create the catsalogue index
	if err = indexService.CreateIndex(ctx, t.client, service.AttributesIndex); err != nil {
		return err
	}

	fmt.Printf("Indexing %d attributes\n", len(t.attributes))

	// Bulk index the attriburtes
	bulk := t.client.Bulk().Index(service.AttributesIndex)
	for _, attr := range t.attributes {
		bulk.Add(elastic.NewBulkIndexRequest().Id(attr.Attribute).Doc(attr))
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
