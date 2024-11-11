package usecase

import (
	"context"

	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

type TaskStore interface {
	Translate(message *domain.Message) *domain.Message
	Search(userMessage *domain.Message) *domain.Message
}

func Translate(ctx context.Context, ts TaskStore, message *domain.Message) *domain.Message {
	return ts.Translate(message)
}

func Search(ctx context.Context, ts TaskStore, message *domain.Message) *domain.Message {
	return ts.Search(message)
}
