package repository

import (
	"database/sql"

	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type TaskStorage struct {
	postgres *sql.DB
	mongo    *mongo.Database
	logger   *zap.SugaredLogger
}

func NewTaskStorage(p *sql.DB, m *mongo.Database, logger *zap.SugaredLogger) *TaskStorage {
	return &TaskStorage{
		postgres: p,
		mongo:    m,
		logger:   logger,
	}
}

func (ts *TaskStorage) Solve(message *domain.Message) {

}

func (ts *TaskStorage) Answer(message *domain.Message) {

}
