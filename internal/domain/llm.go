package domain

type LLMRequest struct {
	Type       int    `json:"type"`
	Prompt     string `json:"prompt"`
	ContextIDs []int  `json:"context_ids"`
}

type ClosestQuestion struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
}

type LLMClosestQuestionsResponse struct {
	ClosestQuestions []ClosestQuestion `json:"closest_questions"`
}

type LLMTheoryResponse struct {
	Response       string `json:"response"`
	UsedContextIDs []int  `json:"used_context_ids"`
}

type LLMProblemResponse struct {
	TRS            string `json:"TRS"`
	Interpretation string `json:"Interpretation"`
}
