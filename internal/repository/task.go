package repository

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

type TaskStorage struct {
	postgres *sql.DB
	mongo    *mongo.Database
	logger   *zap.SugaredLogger
	cfg      *bootstrap.Config
}

func NewTaskStorage(p *sql.DB, m *mongo.Database, logger *zap.SugaredLogger, cfg *bootstrap.Config) *TaskStorage {
	return &TaskStorage{
		postgres: p,
		mongo:    m,
		logger:   logger,
		cfg:      cfg,
	}
}

func (ts *TaskStorage) Solve(message *domain.Message) {

}

func (ts *TaskStorage) Answer(message *domain.Message) {

}

func (ts *TaskStorage) FormAndSendImagesRequest(files []domain.File) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	for _, file := range files {
		openedFile, err := os.Open(file.Path)
		if err != nil {
			fmt.Println("Ошибка при открытии файла:", err)
			return
		}

		// Создаем часть формы для каждого файла
		part, err := writer.CreateFormFile("file", file.Name)
		if err != nil {
			fmt.Println("Ошибка при создании формы:", err)
			openedFile.Close() // Закрываем файл при ошибке
			return
		}

		_, err = io.Copy(part, openedFile)
		if err != nil {
			fmt.Println("Ошибка при копировании файла:", err)
			openedFile.Close()
			return
		}

		openedFile.Close()
	}

	err := writer.WriteField("purpose", "assistants")
	if err != nil {
		fmt.Println("Ошибка при добавлении цели:", err)
		return
	}

	err = writer.Close()
	if err != nil {
		fmt.Println("Ошибка при закрытии writer:", err)
		return
	}

	// Создаем HTTP-запрос
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/files", &requestBody)
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+"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")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка при отправке запроса:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Ошибка при чтении ответа:", err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Файл успешно загружен:", string(body))
	} else {
		fmt.Printf("Ошибка: статус %d\n%s\n", resp.StatusCode, string(body))
	}
}

func (ts *TaskStorage) FormMessageReq(message *tgbotapi.Message) domain.OpenAiResponse {
	openAiReq := domain.OpenAiRequest{Messages: make([]domain.GptMessage, 0)}
	openAiReq.Model = "gpt-4o"

	gptMessageUser := domain.GptMessage{Content: message.Text, Role: "user"}
	gptMessageSystem := domain.GptMessage{Content: "ты помощник в решении задач ТФЯ, отвечаешь очень понятно, развёрнуто, по-русски", Role: "system"}
	openAiReq.Messages = append(openAiReq.Messages, gptMessageSystem)
	openAiReq.Messages = append(openAiReq.Messages, gptMessageUser)

	jsonReq, err := json.Marshal(openAiReq)
	if err != nil {
		ts.logger.Error(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonReq))
	if err != nil {
		ts.logger.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+"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")
	resp, err := client.Do(req)
	if err != nil {
		ts.logger.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ts.logger.Error(err)
	}
	fmt.Println(openAiReq)
	response := ts.ParseGptResponse(body)
	return response
}

func (ts *TaskStorage) ParseGptResponse(jsonResponse []byte) domain.OpenAiResponse {
	response := domain.OpenAiResponse{}
	err := json.Unmarshal(jsonResponse, &response)
	if err != nil {
		ts.logger.Error(err)
	}
	return response
}

func (ts *TaskStorage) SendToOpenAiTask(message *tgbotapi.Message, files []domain.File) domain.OpenAiResponse {
	ts.FormAndSendImagesRequest(files)
	answer := ts.FormMessageReq(message)
	fmt.Println(answer)
	return answer
}

func (ts *TaskStorage) SendImageToOpenAi(message *tgbotapi.Message) {
	//message.Photo[0].FileID
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

func (ts *TaskStorage) Translate(message *domain.Message) *domain.Message {
	translateRequest := TranslateRequest{
		Messages:           make([]string, 0),
		FolderID:           ts.cfg.YandexTranslateFolderId,
		TargetLanguageCode: "en",
	}

	translateRequest.Messages = append(translateRequest.Messages, message.OriginalMessageText)
	translateResponse := ts.SendToYandex(translateRequest)
	message.TranslatedMessageText = translateResponse.Translations[0].Text // TODO fix this hardcode
	return message
}

func (ts *TaskStorage) SendToYandex(request TranslateRequest) (response TranslateResponse) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		ts.logger.Error(err)
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", ts.cfg.YandexTranslateUrl, bytes.NewBuffer(jsonRequest))
	if err != nil {
		ts.logger.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", ts.cfg.YandexKey)
	resp, err := client.Do(req)
	if err != nil {
		ts.logger.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ts.logger.Error(err)
	}
	response = ts.ParseTranslateResponse(body)
	return response
}

func (ts *TaskStorage) ParseTranslateResponse(jsonResponse []byte) (response TranslateResponse) {
	err := json.Unmarshal(jsonResponse, &response)
	if err != nil {
		ts.logger.Error(err)
	}
	return response
}
