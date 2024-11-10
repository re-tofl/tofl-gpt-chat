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

type THandler struct {
	cfg    *bootstrap.Config
	log    *zap.SugaredLogger
	task   *repository.TaskStorage
	openAi *repository.OpenaiStorage
}

func NewTaskHandler(cfg *bootstrap.Config, log *zap.SugaredLogger, task *repository.TaskStorage, open *repository.OpenaiStorage) *THandler {
	return &THandler{
		cfg:  cfg,
		log:  log,
		task: task,
	}
}

func (taskHandler *THandler) SendToOpenAi(ctx context.Context, message *tgbotapi.Message, files []domain.File) {
	usecase.ResolveProblemOnImage(ctx, taskHandler.openAi, taskHandler.log, message, files)
}
