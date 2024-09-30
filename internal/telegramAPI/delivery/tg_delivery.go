package delivery

import (
	"go.uber.org/zap"
	"tgbot/domain"
	tgUsecase "tgbot/internal/telegramAPI/usecase"
)

type TelegramHandler struct {
	logger *zap.SugaredLogger
	tgBot  *domain.Telegram
}

func NewTelegramHandler(logger *zap.SugaredLogger) *TelegramHandler {
	return &TelegramHandler{
		logger: logger,
	}
}

func (tg *TelegramHandler) CreateTelegramBot(tgKey string) {
	bot, err := tgUsecase.CreateNewTgBot(tgKey)
	if err != nil {
		tg.logger.Error(err)
	}
	tg.tgBot = bot
}
