package usecase

import (
	"context"
	"fmt"
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

func (t *TaskUsecase) RateTheory(ctx context.Context, message *tgbotapi.Message) error {
	if message.Text == "+" {
		t.metrics.GoodResponsesLLM.Inc()
		return nil
	}

	if message.Text == "-" {
		t.metrics.BadResponsesLLM.Inc()
		return nil
	}

	return fmt.Errorf("неверный формат")
}
