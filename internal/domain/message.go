package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Message struct {
	SenderChatID          int64
	OriginalMessageText   string   `json:"originalText"`
	TranslatedMessageText string   `json:"translatedText"`
	Context               []bson.M `json:"context"`
	CreatedAt             time.Time
}
