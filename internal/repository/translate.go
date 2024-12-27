package repository

import (
	"database/sql"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
)

type TranslatorStorage struct {
	postgres *sql.DB
	mongo    *mongo.Database
	logger   *zap.SugaredLogger
	cfg      *bootstrap.Config
}
