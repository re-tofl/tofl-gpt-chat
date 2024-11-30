package domain

type DatabaseItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type EmbeddingResp struct {
	Embedding []float64 `json:"embedding"`
}
