package usecase

import (
	"context"
	"fmt"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type RatingRepository interface {
	SaveRating(ctx context.Context, rating domain.Rating) error
}

type TaskUsecase struct {
	metrics    *PrometheusMetrics
	ratingRepo RatingRepository
}

func NewTaskUsecase(ratingRepo RatingRepository) *TaskUsecase {
	return &TaskUsecase{
		metrics:    NewPrometheusMetrics(),
		ratingRepo: ratingRepo,
	}
}

func (t *TaskUsecase) RateTheory(ctx context.Context, message *tgbotapi.Message, contextID string) error {
	ratingValue, err := strconv.ParseFloat(message.Text, 64)
	if err != nil || ratingValue < 1 || ratingValue > 10 {
		return fmt.Errorf("неверное значение: %s. Должно быть от 1 до 10", message.Text)
	}

	t.metrics.ResponseRating.WithLabelValues(message.From.UserName).Observe(ratingValue)
	return t.ratingRepo.SaveRating(ctx, domain.Rating{
		ChatID:  message.Chat.ID,
		Context: contextID,
		Rating:  int(ratingValue),
	})
}
