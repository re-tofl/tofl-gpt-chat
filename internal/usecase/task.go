package usecase

import "github.com/re-tofl/tofl-gpt-chat/internal/domain"

type TaskStore interface {
	Translate(message *domain.Message) *domain.Message
}
