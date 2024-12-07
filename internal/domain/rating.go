package domain

type Rating struct {
	ContextID string `json:"context_id"`
	Rating    int    `json:"rating"`
}
