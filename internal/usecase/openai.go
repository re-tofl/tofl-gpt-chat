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

type OpenAiUseCase struct {
	store OpenAiStore
}

func NewOpenAiUseCase(store OpenAiStore) *OpenAiUseCase {
	return &OpenAiUseCase{
		store: store,
	}
}

func (u *OpenAiUseCase) SaveMedia(ctx context.Context, message *tgbotapi.Message, bot *tgbotapi.BotAPI) string {
	files := u.store.SaveMedia(message, bot)
	gptResponse := u.store.SendPDF(message, files)
	return gptResponse
}

func (u *OpenAiUseCase) SendToGpt(ctx context.Context, message *tgbotapi.Message) string {
	gptResponse := u.store.SendAndGetAnswerFromGptNonFineTuned(message)
	return gptResponse
}
