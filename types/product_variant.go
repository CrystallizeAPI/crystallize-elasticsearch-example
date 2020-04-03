package types

type VariantAttribute struct {
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

type ProductVariant struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	SKU        string             `json:"sku"`
	Price      float32            `json:"price"`
	Stock      int                `json:"stock"`
	IsDefault  bool               `json:"isDefault"`
	Attributes []VariantAttribute `json:"attributes"`
	Images     *[]Image           `json:"images"`
}
