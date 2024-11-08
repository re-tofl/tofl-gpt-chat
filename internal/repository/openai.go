package repository

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"go.uber.org/zap"
)

type OpenaiStorage struct {
	logger *zap.SugaredLogger
	cfg    *bootstrap.Config
}

func NewOpenaiStorage(logger *zap.SugaredLogger, cfg *bootstrap.Config) *OpenaiStorage {
	return &OpenaiStorage{
		logger: logger,
		cfg:    cfg,
	}
}

func (open *OpenaiStorage) PutImagesTo(message *tgbotapi.Message) {

}

func (open *OpenaiStorage) SendRequestToOpenAI(message *tgbotapi.Message) {
	//message.Photo[0].FileID
}
