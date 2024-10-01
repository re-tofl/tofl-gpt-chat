package repository

import (
	"database/sql"

	"go.uber.org/zap"
	"tgbot/domain"
)

type ChatStorage struct {
	logger *zap.SugaredLogger
	db     *sql.DB
}

func NewChatStorage(Logger *zap.SugaredLogger, db *sql.DB) *ChatStorage {
	return &ChatStorage{
		logger: Logger,
		db:     db,
	}
}

func (c *ChatStorage) FullTextSearch(originalMessageText string) []domain.Chunk {
	return make([]domain.Chunk, 0)
}
