package domain

type Rating struct {
	ChatID  int64  `json:"chat_id"`
	Context string `json:"context"`
	Rating  int    `json:"rating"`
}
