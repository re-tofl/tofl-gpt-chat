package delivery

import (
	"database/sql"
	"time"

	"go.uber.org/zap"
	"tgbot/domain"
	chatRepo "tgbot/internal/chat/repository"
	chatUsecase "tgbot/internal/chat/usecase"
)

type ChatHandler struct {
	Logger *zap.SugaredLogger
	TgBot  *domain.Telegram
	Chat   *chatRepo.ChatStorage
}

func NewChatHandler(logger *zap.SugaredLogger, db *sql.DB) *ChatHandler {
	return &ChatHandler{
		Logger: logger,
		Chat:   chatRepo.NewChatStorage(logger, db),
	}
}

func (chat *ChatHandler) SendMessageToLLM(chatID int64, text string) {
	message := domain.Message{
		SenderChatID:        chatID,
		OriginalMessageText: text,
		CreatedAt:           time.Now(),
		MessageContext:      make([]domain.Chunk, 0),
	}
	chatUsecase.SendMessageFromUserToLLM(chat.Chat, message)
}
