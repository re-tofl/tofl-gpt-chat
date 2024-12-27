package domain

type LLMRequest struct {
	Type   int    `json:"type"`
	Prompt string `json:"prompt"`
}

type LLMTheoryResponse struct {
	Response  string `json:"response"`
	ContextID int    `json:"context_id"`
}
