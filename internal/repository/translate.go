package repository

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type TranslatorStorage struct {
	postgres *sql.DB
	mongo    *mongo.Database
	logger   *zap.SugaredLogger
	cfg      *bootstrap.Config
}

type TranslateRequest struct {
	Messages           []string `json:"texts"`
	FolderID           string   `json:"folderId"`
	TargetLanguageCode string   `json:"targetLanguageCode"`
}

type TranslateResponse struct {
	Translations []Translations `json:"translations"`
}

type Translations struct {
	Text                 string `json:"text"`
	DetectedLanguageCode string `json:"detectedLanguageCode"`
}

func NewTranslatorStorage(p *sql.DB, m *mongo.Database, logger *zap.SugaredLogger, cfg *bootstrap.Config) *TranslatorStorage {
	return &TranslatorStorage{
		postgres: p,
		mongo:    m,
		logger:   logger,
		cfg:      cfg,
	}
}

func (t *TranslatorStorage) Translate(message *domain.Message) *domain.Message {
	translateRequest := TranslateRequest{
		Messages:           make([]string, 0),
		FolderID:           t.cfg.YandexTranslateFolderId,
		TargetLanguageCode: "en",
	}

	translateRequest.Messages = append(translateRequest.Messages, message.OriginalMessageText)
	translateResponse := t.SendToYandex(translateRequest)
	message.TranslatedMessageText = translateResponse.Translations[0].Text // TODO fix this hardcode
	return message

}

func (t *TranslatorStorage) SendToYandex(request TranslateRequest) (response TranslateResponse) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		t.logger.Error(err)
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", t.cfg.YandexTranslateUrl, bytes.NewBuffer(jsonRequest))
	if err != nil {
		t.logger.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", t.cfg.YandexKey)
	resp, err := client.Do(req)
	if err != nil {
		t.logger.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.logger.Error(err)
	}
	response = t.ParseTranslateResponse(body)
	return response
}

func (t *TranslatorStorage) ParseTranslateResponse(jsonResponse []byte) (response TranslateResponse) {
	err := json.Unmarshal(jsonResponse, &response)
	if err != nil {
		t.logger.Error(err)
	}
	return response
}
