package types

import "github.com/CrystallizeAPI/crystallize-go-types/types"

// ElasticProduct represents a product as it is stored in elasticsearch.
type ElasticProduct struct {
	Variant types.ProductVariant `json:"variant"`
	Product types.Product        `json:"product"`
}
