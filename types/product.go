package types

type Product struct {
	CatalogueItem
	Variants []ProductVariant `json:"variants"`
}
