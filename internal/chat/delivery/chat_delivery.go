package delivery

import (
	"go.uber.org/zap"
	"tgbot/domain"
	tgUsecase "tgbot/internal/telegramAPI/usecase"
)

type ChatHandler struct {
	Logger *zap.SugaredLogger
	TgBot  *domain.Telegram
}

func NewChatHandler(logger *zap.SugaredLogger) *ChatHandler {
	return &ChatHandler{
		Logger: logger,
	}
}

func (chat *ChatHandler) SendMessageTo(tgKey string) {
	bot, err := tgUsecase.CreateNewTgBot(tgKey)
	if err != nil {
		chat.Logger.Error(err)
	}
	chat.TgBot = bot
}
