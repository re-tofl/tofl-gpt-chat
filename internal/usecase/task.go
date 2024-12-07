package usecase

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TaskUsecase struct {
	metrics *PrometheusMetrics
}

func NewTaskUsecase() *TaskUsecase {
	return &TaskUsecase{
		metrics: NewPrometheusMetrics(),
	}
}

func (t *TaskUsecase) RateTheory(ctx context.Context, message *tgbotapi.Message, contextID string) error {
	ratingValue, err := strconv.ParseFloat(message.Text, 64)
	if err != nil || ratingValue < 1 || ratingValue > 10 {
		return fmt.Errorf("неверное значение: %s. Должно быть от 1 до 10", message.Text)
	}

	t.metrics.ResponseRating.WithLabelValues(message.From.UserName).Observe(ratingValue)
	return nil
}
