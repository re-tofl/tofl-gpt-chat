package repository

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

type OpenaiStorage struct {
	logger      *zap.SugaredLogger
	cfg         *bootstrap.Config
	fileIds     []string
	assistantId string
}

func NewOpenaiStorage(logger *zap.SugaredLogger, cfg *bootstrap.Config) *OpenaiStorage {
	return &OpenaiStorage{
		logger:  logger,
		cfg:     cfg,
		fileIds: make([]string, 0),
	}
}

type OpenAiFileResponse struct {
	Response string `json:"response"`
}

func (open *OpenaiStorage) ProcessFilesAndSendRequest(message *tgbotapi.Message, files []domain.File) OpenAiFileResponse {
	fileResp := OpenAiFileResponse{}
	convertedFiles := []string{}

	for _, file := range files {
		fileExt := strings.ToLower(filepath.Ext(file.Path)[1:]) // Получаем расширение без точки
		fmt.Println("FileExt:", fileExt)
		var err error

		if fileExt == "pdf" {
			cmd := exec.Command("pdftoppm", file.Path, "output", "-png")
			err = cmd.Run()
			if err != nil {
				fmt.Println("Ошибка при конвертации PDF в изображения:", err)
				continue // Продолжаем обработку других файлов
			}

			convertedFiles, err = filepath.Glob("output-*.png")
			if err != nil {
				open.logger.Error("Ошибка при поиске PNG файлов:", err)
				continue
			}
		} else if fileExt == "jpg" || fileExt == "jpeg" {
			fmt.Println("here")
			convertedFiles, err = filepath.Glob("upload/*.jpg")
			if err != nil {
				open.logger.Error("Ошибка при поиске JPG файлов:", err)
				continue
			}
		}

		if len(convertedFiles) == 0 {
			fmt.Println("Нет файлов для обработки.")
			continue
		}

		jsonReq := domain.OpenAiImageRequest{
			Base64: make([]domain.Bases, 0),
			Prompt: message.Text,
		}

		for _, imgFile := range convertedFiles {
			imageData, err := ioutil.ReadFile(imgFile)
			if err != nil {
				fmt.Printf("Ошибка при чтении файла %s: %v\n", imgFile, err)
				continue
			}

			encoded := base64.StdEncoding.EncodeToString(imageData)
			jsonReq.Base64 = append(jsonReq.Base64, domain.Bases{Base64: encoded})
		}

		byteReq, err := json.Marshal(jsonReq)
		if err != nil {
			open.logger.Error("Ошибка при маршализации запроса:", err)
			continue
		}

		req, err := http.NewRequest("POST", "http://localhost:8085/image", bytes.NewBuffer(byteReq))
		if err != nil {
			open.logger.Error("Ошибка при создании запроса:", err)
			continue
		}
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Ошибка при отправке запроса:", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Ошибка при чтении ответа:", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			err = json.Unmarshal(body, &fileResp)
			if err != nil {
				open.logger.Error("Ошибка при разборе ответа:", err)
			}
			fmt.Println("Ответ:", string(body), fileResp)
		} else {
			open.logger.Error("Статус ошибки:", resp.StatusCode, string(body))
		}
	}
	return fileResp
}

func (open *OpenaiStorage) SendPDF(message *tgbotapi.Message, files []domain.File) string {
	chatGptResponse := open.ProcessFilesAndSendRequest(message, files)
	if chatGptResponse.Response != "" {
		return chatGptResponse.Response
	}
	return "chatGpt returns err!"
}
