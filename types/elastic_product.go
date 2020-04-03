package types

// ElasticProduct represents a product as it is stored in elasticsearch. This
type ElasticProduct struct {
	Variant ProductVariant `json:"variant"`
	Product Product        `json:"product"`
}
