package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"github.com/re-tofl/tofl-gpt-chat/internal/adapters"
	"github.com/re-tofl/tofl-gpt-chat/pkg/interr"
)

type UserRepository struct {
	db *adapters.AdapterPG
}

func NewUserRepository(db *adapters.AdapterPG) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (u *UserRepository) SetState(ctx context.Context, chatID int64, state string) error {
	query := `INSERT INTO chat.user(chat_id, state) VALUES($1, $2) 
              ON CONFLICT(chat_id) DO UPDATE SET state = $2`

	_, err := u.db.Exec(ctx, query, state, chatID)
	if err != nil {
		return interr.NewInternalError(err, "UserRepository.SetState db.Exec")
	}

	return nil
}

func (u *UserRepository) GetState(ctx context.Context, chatID int64) (string, error) {
	query := "SELECT state FROM chat.user WHERE chat_id = $1"

	var state string

	err := u.db.QueryRow(ctx, query, chatID).Scan(&state)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", interr.NewNotFoundError(err, "UserRepository.GetState db.QueryRow")
	}

	if err != nil {
		return "", interr.NewInternalError(err, "UserRepository.GetState db.QueryRow")
	}

	return state, nil
}
