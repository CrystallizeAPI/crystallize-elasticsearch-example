package types

// CatalogueItem represents an item returned by the catalogue. This includes
// products, documents, and folders.
type CatalogueItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
