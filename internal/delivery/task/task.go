package task

import (
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/repository"
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
