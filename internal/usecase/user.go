package usecase

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

// TODO: отрефакторить код, ввести контексты, вынести обработку ошибок в юзкейсы

type UserRepository interface {
	SetState(ctx context.Context, chatID int64, state string)
	GetState(ctx context.Context, chatID int64) string
	CheckAccExists(ctx context.Context, chatID int64) (bool, domain.User)
	Register(ctx context.Context, user *domain.User) int
}

type UserHandler struct {
	log     *zap.SugaredLogger
	repo    UserRepository
	metrics *PrometheusMetrics
}

func NewUserHandler(Logger *zap.SugaredLogger, repo UserRepository) *UserHandler {
	return &UserHandler{
		log:     Logger,
		repo:    repo,
		metrics: NewPrometheusMetrics(),
	}
}

func (u *UserHandler) SetUserState(chatID int64, state string) {
	u.repo.SetState(context.TODO(), chatID, state)
	u.metrics.Methods.WithLabelValues("SwitchState").Inc()
}

func (u *UserHandler) GetUserState(chatID int64) string {
	state := u.repo.GetState(context.TODO(), chatID)
	u.metrics.Methods.WithLabelValues("GetState").Inc()

	return state
}

func (u *UserHandler) CheckAccExists(chatID int64) (bool, domain.User) {
	u.metrics.Methods.WithLabelValues("CheckAcc").Inc()
	exists, user := u.repo.CheckAccExists(context.TODO(), chatID)

	return exists, user
}

func (u *UserHandler) RegisterUser(message tgbotapi.Message) domain.User {
	user := domain.User{}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.ChatID = message.Chat.ID
	user.State = "start"
	user.Nickname = message.Chat.UserName

	user.Id = u.repo.Register(context.TODO(), &user)
	u.metrics.RegisteredUserCount.Inc()

	return user
}
