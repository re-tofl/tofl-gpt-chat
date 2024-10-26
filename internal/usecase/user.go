package usecase

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/pkg/interr"
)

// TODO: отрефакторить код, ввести контексты, вынести обработку ошибок в юзкейсы

type UserRepository interface {
	SetState(ctx context.Context, chatID int64, state string) error
	GetState(ctx context.Context, chatID int64) (string, error)
}

type UserUsecase struct {
	log     *zap.SugaredLogger
	repo    UserRepository
	metrics *PrometheusMetrics
}

func NewUserHandler(logger *zap.SugaredLogger, repo UserRepository) *UserUsecase {
	return &UserUsecase{
		log:     logger,
		repo:    repo,
		metrics: NewPrometheusMetrics(),
	}
}

func (u *UserUsecase) SetUserState(ctx context.Context, chatID int64, state string) {
	err := u.repo.SetState(ctx, chatID, state)
	u.metrics.Methods.WithLabelValues("SwitchState").Inc()

	if errors.Is(err, interr.ErrNotFound) {
		u.log.Errorw("UserUsecase.SetUserState repo.SetState", zap.Error(err))
	}
}

func (u *UserUsecase) GetUserState(ctx context.Context, chatID int64) string {
	state, err := u.repo.GetState(ctx, chatID)
	u.metrics.Methods.WithLabelValues("GetState").Inc()

	if errors.Is(err, interr.ErrNotFound) {
		u.log.Errorw("UserUsecase.GetUserState repo.GetState", zap.Error(err))
	}

	return state
}
