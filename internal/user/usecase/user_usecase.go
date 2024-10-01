package usecase

import (
	"context"

	"tgbot/domain"
)

type UserStore interface {
	SetState(ctx context.Context, chatID int64, state string)
	GetState(ctx context.Context, chatID int64) string
	CheckAccExists(ctx context.Context, chatID int64) (bool, domain.User)
	Register(ctx context.Context, user *domain.User) int
}

func GetUserState(chatID int64, us UserStore) string {
	ctx := context.Background()
	userState := us.GetState(ctx, chatID)
	return userState
}

func SetUserState(chatID int64, state string, us UserStore) {
	ctx := context.Background()
	us.SetState(ctx, chatID, state)
}

func CheckAccExists(chatID int64, us UserStore) (bool, domain.User) {
	ctx := context.Background()
	exists, user := us.CheckAccExists(ctx, chatID)
	return exists, user
}

func Register(user *domain.User, us UserStore) int {
	ctx := context.Background()
	userID := us.Register(ctx, user)
	return userID
}
