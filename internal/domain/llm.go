package domain

type LLMRequest struct {
	Type   int    `json:"type"`
	Prompt string `json:"prompt"`
}

type LLMTheoryResponse struct {
	Result string `json:"result"`
}
