package domain

import "time"

type Message struct {
	SenderChatID        int64
	OriginalMessageText string
	MessageContext      []Chunk
	CreatedAt           time.Time
}

type Chunk struct {
	key   string
	value string
}
