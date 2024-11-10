package usecase

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

type OpenAiStore interface {
	SendPDF(message *tgbotapi.Message, files []domain.File) string
}

func ResolveProblemOnImage(ctx context.Context, os OpenAiStore, logger *zap.SugaredLogger, message *tgbotapi.Message, files []domain.File) string {
	gptResponse := os.SendPDF(message, files)
	return gptResponse

}
