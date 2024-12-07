package domain

type LLMRequest struct {
	Type   int    `json:"type"`
	Prompt string `json:"prompt"`
}

type LLMTheoryResponse struct {
	Response  string `json:"response"`
	ContextID string `json:"context_id"`
}
