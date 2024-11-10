package usecase

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"go.uber.org/zap"
)

type OpenAiStore interface {
	SendToOpenAi(message *tgbotapi.Message, files []domain.File) domain.OpenAiResponse
}

func ResolveProblemOnImage(ctx context.Context, os OpenAiStore, logger *zap.SugaredLogger, message *tgbotapi.Message, files []domain.File) {
	os.SendToOpenAi(message, files)
}
