package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

type SpeechStorage struct {
	logger *zap.SugaredLogger
	cfg    *bootstrap.Config
}

func NewSpeechStorage(logger *zap.SugaredLogger, cfg *bootstrap.Config) *SpeechStorage {
	return &SpeechStorage{
		logger: logger,
		cfg:    cfg,
	}
}

func (s *SpeechStorage) SpeechToText(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		s.logger.Error(err)
		return ""
	}
	defer file.Close()

	ConvertToVorbis(filePath)

	preparedResp := s.CreateYandexSpeechToTextResponse(filePath)
	if preparedResp == nil {
		return ""
	}
	defer preparedResp.Body.Close()

	speechToText := s.HandleResponse(preparedResp)

	return speechToText.Result
}

func (s *SpeechStorage) CreateYandexSpeechToTextResponse(filePath string) *http.Response {
	audioData, err := os.ReadFile(filePath)
	if err != nil {
		s.logger.Error("Ошибка чтения файла:", err)
		return nil
	}

	baseURL := "https://stt.api.cloud.yandex.net/speech/v1/stt:recognize"
	u, err := url.Parse(baseURL)
	if err != nil {
		s.logger.Error("Ошибка парсинга URL:", err)
		return nil
	}

	query := u.Query()
	query.Set("lang", "ru-RU")
	query.Set("folderId", "b1gq4i9e5unl47m0kj5f")
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(audioData))
	if err != nil {
		s.logger.Error("Ошибка создания HTTP-запроса:", err)
		return nil
	}

	req.Header.Set("Authorization", s.cfg.YandexKey)
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Ошибка выполнения HTTP-запроса:", err)
		return nil
	}

	return resp
}

func ConvertToVorbis(inputFile string) {
	inputPath := inputFile

	tempFile := inputPath + ".tmp"

	cmd := exec.Command("ffmpeg", "-i", inputPath, "-c:a", "libvorbis", "-b:a", "128k", "-f", "ogg", tempFile)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("Ошибка при конвертации:", err)
		fmt.Println("Вывод ffmpeg:", stderr.String())
		return
	}

	err = os.Rename(tempFile, inputPath)
	if err != nil {
		fmt.Println("Ошибка при замене оригинального файла:", err)
		return
	}

	fmt.Println("Файл успешно конвертирован:", inputPath)
}

func (s *SpeechStorage) HandleResponse(yandexResp *http.Response) domain.SpeechToTextResponse {
	respStruct := domain.SpeechToTextResponse{}

	jsonResp, err := io.ReadAll(yandexResp.Body)
	if err != nil {
		s.logger.Error("Ошибка чтения ответа:", err)
		return respStruct
	}
	s.logger.Infof("Ответ от Yandex: %s", string(jsonResp))

	err = json.Unmarshal(jsonResp, &respStruct)
	if err != nil {
		s.logger.Error("Ошибка разбора JSON:", err)
	}

	return respStruct
}
