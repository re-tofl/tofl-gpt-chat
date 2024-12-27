package usecase

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

type OpenAiStore interface {
	SendPDF(message *tgbotapi.Message, files []domain.File) string
	SaveMedia(message *tgbotapi.Message, bot *tgbotapi.BotAPI) []domain.File
	SendAndGetAnswerFromGptNonFineTuned(message *tgbotapi.Message) string
}

func SaveMedia(ctx context.Context, os OpenAiStore, message *tgbotapi.Message, bot *tgbotapi.BotAPI) string {
	files := os.SaveMedia(message, bot)
	gptResponse := os.SendPDF(message, files)
	return gptResponse
}

func SendToGpt(ctx context.Context, os OpenAiStore, message *tgbotapi.Message) string {
	gptResponse := os.SendAndGetAnswerFromGptNonFineTuned(message)
	return gptResponse
}
