package domain

type LLMRequest struct {
	Context string `json:"context"`
}

type LLMTheoryResponse struct {
	Result string `json:"result"`
}
