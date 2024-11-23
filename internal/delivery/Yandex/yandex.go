package Yandex

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/repository"
	"github.com/re-tofl/tofl-gpt-chat/internal/usecase"
	"go.uber.org/zap"
)

type SpeechHandler struct {
	cfg    *bootstrap.Config
	log    *zap.SugaredLogger
	speech *repository.SpeechStorage
}

func NewSpeechHandler(cfg *bootstrap.Config, log *zap.SugaredLogger, speech *repository.SpeechStorage) *SpeechHandler {
	return &SpeechHandler{
		cfg:    cfg,
		log:    log,
		speech: speech,
	}
}

func (sh *SpeechHandler) SpeechToText(ctx context.Context, voiceFilePath string, bot *tgbotapi.BotAPI) string {
	text := usecase.ConvertSpeechToText(ctx, sh.speech, voiceFilePath, bot)
	return text
}
