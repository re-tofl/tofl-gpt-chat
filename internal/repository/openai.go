package repository

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

type OpenAiResponse struct {
	Response string `json:"response"`
}

func (open *OpenaiStorage) ProcessFilesAndSendRequest(message *tgbotapi.Message, files []domain.File) OpenAiResponse {
	fileResp := OpenAiResponse{}

	for _, file := range files {
		fileExt := getFileExtension(file.Path)
		var err error
		var convertedFiles []string

		switch fileExt {
		case "pdf":
			convertedFiles, err = convertPDFToImages(file.Path)
		case "jpg", "jpeg":
			convertedFiles, err = findJPGFiles("upload/*.jpg")
		default:
			continue
		}

		if err != nil {
			open.logger.Error("Ошибка при обработке файла:", err)
			return fileResp
		}

		jsonReq, err := createJSONRequest(message.Text, convertedFiles)
		if err != nil {
			open.logger.Error("Ошибка при создании JSON запроса:", err)
			return fileResp
		}

		respBody, err := sendHTTPRequest("http://66.151.42.116:8085/image", jsonReq)
		if err != nil {
			open.logger.Error("Ошибка при отправке HTTP запроса:", err)
			return fileResp
		}

		err = json.Unmarshal(respBody, &fileResp)
		if err != nil {
			open.logger.Error("Ошибка при разборе ответа:", err)
		}
	}

	return fileResp
}

func getFileExtension(filePath string) string {
	return strings.ToLower(filepath.Ext(filePath)[1:])
}

func findJPGFiles(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func convertPDFToImages(filePath string) ([]string, error) {
	cmd := exec.Command("pdftoppm", filePath, "output", "-png")
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ошибка при конвертации PDF в изображения: %w", err)
	}
	return filepath.Glob("output-*.png")
}

func createJSONRequest(prompt string, imageFiles []string) ([]byte, error) {
	jsonReq := domain.OpenAiImageRequest{
		Base64: make([]domain.Bases, 0),
		Prompt: prompt,
	}

	for _, imgFile := range imageFiles {
		imageData, err := ioutil.ReadFile(imgFile)
		if err != nil {
			return nil, fmt.Errorf("ошибка при чтении файла %s: %w", imgFile, err)
		}
		encoded := base64.StdEncoding.EncodeToString(imageData)
		jsonReq.Base64 = append(jsonReq.Base64, domain.Bases{Base64: encoded})
	}

	return json.Marshal(jsonReq)
}

func sendHTTPRequest(url string, requestBody []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании HTTP запроса: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при отправке запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("статус ошибки: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (open *OpenaiStorage) SendPDF(message *tgbotapi.Message, files []domain.File) string {
	chatGptResponse := open.ProcessFilesAndSendRequest(message, files)
	if chatGptResponse.Response != "" {
		return chatGptResponse.Response
	}
	return "chatGpt returns err!"
}

func (open *OpenaiStorage) SaveMedia(message *tgbotapi.Message, bot *tgbotapi.BotAPI) []domain.File {
	files := make([]domain.File, 0)

	if message.Photo != nil {
		photoFiles, err := open.savePhotoFiles(message.Photo, bot)
		if err != nil {
			open.logger.Error("Ошибка при обработке фото:", err)
			return files
		}
		files = append(files, photoFiles...)
	}

	if message.Document != nil {
		documentFile, err := open.saveDocumentFile(message.Document, bot)
		if err != nil {
			open.logger.Error("Ошибка при обработке документа:", err)
			return files
		}
		files = append(files, documentFile)
	}

	return files
}

func (open *OpenaiStorage) savePhotoFiles(photos *[]tgbotapi.PhotoSize, bot *tgbotapi.BotAPI) ([]domain.File, error) {
	files := make([]domain.File, 0)
	images := *photos
	fileID := images[len(images)-1].FileID // выбираем изображение наилучшего качества

	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении файла: %w", err)
	}

	filePath, err := open.downloadAndSaveFile(file, fmt.Sprintf("image_%s.jpg", fileID))
	if err != nil {
		return nil, err
	}

	files = append(files, domain.File{Name: filepath.Base(filePath), Path: filePath})
	return files, nil
}

func (open *OpenaiStorage) saveDocumentFile(document *tgbotapi.Document, bot *tgbotapi.BotAPI) (domain.File, error) {
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: document.FileID})
	if err != nil {
		return domain.File{}, fmt.Errorf("ошибка при получении документа: %w", err)
	}

	fileExtension := getFileExtension(document.FileName)
	if fileExtension == "" {
		fileExtension = filepath.Ext(file.FilePath)
	}

	if fileExtension == "" {
		fileExtension = ".txt"
	}

	fileName := fmt.Sprintf("%s_%d%s", document.FileName, time.Now().UnixNano(), fileExtension)
	filePath, err := open.downloadAndSaveFile(file, fileName)
	if err != nil {
		return domain.File{}, err
	}

	return domain.File{Name: fileName, Path: filePath}, nil
}

func (open *OpenaiStorage) downloadAndSaveFile(file tgbotapi.File, fileName string) (string, error) {
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", open.cfg.TGBotToken, file.FilePath)

	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("ошибка при загрузке файла: %w", err)
	}
	defer resp.Body.Close()

	filePath := filepath.Join("upload", fileName)
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании файла: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при сохранении файла: %w", err)
	}

	return filePath, nil
}

func (open *OpenaiStorage) SendAndGetAnswerFromGptNonFineTuned(message *tgbotapi.Message) string {
	req := domain.OpenAiTextRequest{
		Prompt: message.Text,
	}
	jsonReq, err := json.Marshal(req)
	if err != nil {
		open.logger.Error(err)
	}
	jsonResp, err := sendHTTPRequest("http://66.151.42.116:8085/text", jsonReq)
	if err != nil {
		open.logger.Error(err)
	}
	gptResp := OpenAiResponse{}
	err = json.Unmarshal(jsonResp, &gptResp)
	if err != nil {
		open.logger.Error(err)
	}
	return gptResp.Response
}
