package domain

type OpenAiRequest struct {
	Model    string       `json:"model"`
	Messages []GptMessage `json:"messages"`
}

type GptMessage struct {
	Role        string       `json:"role"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type File struct {
	Name string `json:"file_name"`
	Path string `json:"file_path"`
}

type OpenAiResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint"`
}

type Choice struct {
	Index        int       `json:"index"`
	Message      GptAnswer `json:"message"`
	Logprobs     *int      `json:"logprobs"`
	FinishReason string    `json:"finish_reason"`
}

type GptAnswer struct {
	Role    string  `json:"role"`
	Content string  `json:"content"`
	Refusal *string `json:"refusal"`
}

type Usage struct {
	PromptTokens            int                    `json:"prompt_tokens"`
	CompletionTokens        int                    `json:"completion_tokens"`
	TotalTokens             int                    `json:"total_tokens"`
	PromptTokensDetails     TokensDetails          `json:"prompt_tokens_details"`
	CompletionTokensDetails CompletionTokensDetail `json:"completion_tokens_details"`
}

type TokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
	AudioTokens  int `json:"audio_tokens"`
}

type CompletionTokensDetail struct {
	ReasoningTokens          int `json:"reasoning_tokens"`
	AudioTokens              int `json:"audio_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

type Attachment struct {
	FileID string `json:"file_id"`
	Tools  []Tool `json:"tools"`
}

type Tool struct {
	Type string `json:"type"`
}

type OpenAiImageRequest struct {
	Base64 []Bases `json:"images"`
	Prompt string  `json:"prompt"`
}

type Bases struct {
	Base64 string `json:"image_base64"`
}
