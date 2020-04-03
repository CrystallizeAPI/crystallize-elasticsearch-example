package types

type ImageVariant struct {
	Key   string `json:"key"`
	URL   string `json:"url"`
	Width int    `json:"width"`
}

type Image struct {
	Key      string         `json:"key"`
	URL      string         `json:"url"`
	Variants []ImageVariant `json:"variants"`
}
