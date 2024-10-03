package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"tgbot/domain"
)

type UserStorage struct {
	logger *zap.SugaredLogger
	db     *sql.DB
}

func NewUserStorage(Logger *zap.SugaredLogger, db *sql.DB) *UserStorage {
	return &UserStorage{
		logger: Logger,
		db:     db,
	}
}

func (u *UserStorage) SetState(ctx context.Context, chatID int64, state string) {
	query := "UPDATE chat.user SET state = $1 WHERE chat_id = $2"
	_, err := u.db.ExecContext(ctx, query, state, chatID)
	if err != nil {
		u.logger.Error("userState updating err: ", err)
	}
}

func (u *UserStorage) GetState(ctx context.Context, chatID int64) string {
	query := "SELECT state FROM chat.user WHERE chat_id = $1"
	fmt.Println(chatID)
	var state string
	err := u.db.QueryRowContext(ctx, query, chatID).Scan(&state)
	if err != nil {
		u.logger.Error("userState getting err: ", err)
	}
	return state
}

func (u *UserStorage) CheckAccExists(ctx context.Context, chatID int64) (bool, domain.User) {
	query := "SELECT id, created_at, updated_at, nickname, chat_id, state FROM chat.user WHERE chat_id = $1"
	var user domain.User
	err := u.db.QueryRowContext(ctx, query, chatID).Scan(&user.Id, &user.CreatedAt, &user.UpdatedAt, &user.Nickname, &user.ChatID, &user.State)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, domain.User{}
		} else {
			u.logger.Error("userState getting err: ", err)
		}
	}
	return true, user
}

func (u *UserStorage) Register(ctx context.Context, user *domain.User) int {
	var userID int
	query := "INSERT INTO chat.user (created_at, updated_at, nickname, chat_id, state) VALUES ($1, $2, $3, $4, $5) returning id"
	err := u.db.QueryRowContext(ctx, query, user.CreatedAt, user.UpdatedAt, user.Nickname, user.ChatID, user.State).Scan(&userID)
	if err != nil {
		u.logger.Error("Register user database err: ", err.Error())
	}
	return userID
}
