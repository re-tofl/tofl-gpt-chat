package domain

type LLMRequest struct {
	Type   int    `json:"type"`
	Prompt string `json:"prompt"`
}

type LLMTheoryResponse struct {
	Response string `json:"response"`
	Context  string `json:"context"`
}
