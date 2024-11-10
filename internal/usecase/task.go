package usecase

import (
	"context"

	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"go.uber.org/zap"
)

type TaskStore interface {
	Translate(message *domain.Message) *domain.Message
	//SendToOpenAi(message *tgbotapi.Message, files []domain.File) domain.OpenAiResponse
}

func Translate(ctx context.Context, ts TaskStore, logger *zap.SugaredLogger, message *domain.Message) *domain.Message {
	return ts.Translate(message)
}

/*func ResolveProblemOnImage(ctx context.Context, ts TaskStore, logger *zap.SugaredLogger, message *tgbotapi.Message, files []domain.File) {
	ts.SendToOpenAi(message, files)
}*/
