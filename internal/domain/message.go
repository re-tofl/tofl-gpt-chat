package domain

import "time"

type Message struct {
	SenderChatID        int64
	OriginalMessageText string
	CreatedAt           time.Time
}
