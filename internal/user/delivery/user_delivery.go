package delivery

import (
	"database/sql"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
	"tgbot/domain"
	userRepo "tgbot/internal/user/repository"
	userUsecase "tgbot/internal/user/usecase"
)

type UserHandler struct {
	Logger *zap.SugaredLogger
	users  *userRepo.UserStorage
}

func NewUserHandler(Logger *zap.SugaredLogger, db *sql.DB) *UserHandler {
	return &UserHandler{
		Logger: Logger,
		users:  userRepo.NewUserStorage(Logger, db),
	}
}

func (u *UserHandler) SetUserState(chatID int64, state string) {
	userUsecase.SetUserState(chatID, state, u.users)
}

func (u *UserHandler) GetUserState(chatID int64) string {
	userState := userUsecase.GetUserState(chatID, u.users)
	return userState
}

func (u *UserHandler) CheckAccExists(chatID int64) (bool, domain.User) {
	exists, user := userUsecase.CheckAccExists(chatID, u.users)
	return exists, user
}

func (u *UserHandler) RegisterUser(message tgbotapi.Message) domain.User {
	user := domain.User{}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.ChatID = message.Chat.ID
	user.State = "start"
	user.Nickname = message.Chat.UserName

	user.Id = userUsecase.Register(&user, u.users)

	return user
}
