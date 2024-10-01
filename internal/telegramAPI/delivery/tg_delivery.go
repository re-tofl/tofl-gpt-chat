package delivery

import (
	"go.uber.org/zap"
	"tgbot/domain"
	tgUsecase "tgbot/internal/telegramAPI/usecase"
)

type TelegramHandler struct {
	Logger *zap.SugaredLogger
	TgBot  *domain.Telegram
}

func NewTelegramHandler(logger *zap.SugaredLogger) *TelegramHandler {
	return &TelegramHandler{
		Logger: logger,
	}
}

func (tg *TelegramHandler) CreateTelegramBot(tgKey string) {
	bot, err := tgUsecase.CreateNewTgBot(tgKey)
	if err != nil {
		tg.Logger.Error(err)
	}
	tg.TgBot = bot
}
