package usecase

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type SpeechStore interface {
	SpeechToText(filePath string) string
}

func ConvertSpeechToText(ctx context.Context, s SpeechStore, filePath string, bot *tgbotapi.BotAPI) string {
	answer := s.SpeechToText(filePath)
	return answer
}
