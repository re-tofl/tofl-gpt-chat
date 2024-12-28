package domain

type Rating struct {
	ChatID         int64 `json:"chat_id"`
	UsedContextIDs []int `json:"used_context_ids"`
	Rating         int   `json:"rating"`
}
