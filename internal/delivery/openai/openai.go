package task

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"github.com/re-tofl/tofl-gpt-chat/internal/repository"
	"github.com/re-tofl/tofl-gpt-chat/internal/usecase"
)

type OpenHandler struct {
	cfg    *bootstrap.Config
	log    *zap.SugaredLogger
	openAi *repository.OpenaiStorage
}

func NewOpenHandler(cfg *bootstrap.Config, log *zap.SugaredLogger) *OpenHandler {
	return &OpenHandler{
		cfg:    cfg,
		log:    log,
		openAi: repository.NewOpenaiStorage(log, cfg),
	}
}

func (openHandler *OpenHandler) SendToOpenAi(ctx context.Context, message *tgbotapi.Message, files []domain.File) {
	openHandler.log.Infof(usecase.ResolveProblemOnImage(ctx, openHandler.openAi, openHandler.log, message, files))
	openHandler.log.Infof("Succesful!")
}