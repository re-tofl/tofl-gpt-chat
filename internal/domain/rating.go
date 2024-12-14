package domain

type Rating struct {
	ChatID    int64 `json:"chat_id"`
	ContextID int   `json:"context_id"`
	Rating    int   `json:"rating"`
}
