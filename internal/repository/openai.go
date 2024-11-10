package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
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

func (open *OpenaiStorage) FormAndSendImagesRequest(files []domain.File) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	for _, file := range files {
		openedFile, err := os.Open(file.Path)
		////// НАЧАТЬ ТУТ
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
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := http.Client{}

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

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var uploadResponse struct {
			ID string `json:"id"`
		}
		err = json.Unmarshal(body, &uploadResponse)
		if err != nil {
			fmt.Errorf("Ошибка при разборе ответа загрузки: %v", err)
		}
		open.fileIds = append(open.fileIds, uploadResponse.ID)
		fmt.Println("Файл успешно загружен:", string(body))
	} else {
		fmt.Errorf("Ошибка загрузки файла: статус %d\n%s", resp.StatusCode, string(body))
	}
}

func (open *OpenaiStorage) FormMessageReq(message *tgbotapi.Message) domain.OpenAiResponse {
	openAiReq := domain.OpenAiRequest{Messages: make([]domain.GptMessage, 0)}
	openAiReq.Model = "gpt-4o"

	gptMessageUser := domain.GptMessage{Content: message.Text, Role: "user"}
	gptMessageSystem := domain.GptMessage{Content: "ты помощник в решении задач ТФЯ, отвечаешь очень понятно, развёрнуто, по-русски", Role: "system"}
	openAiReq.Messages = append(openAiReq.Messages, gptMessageSystem)
	openAiReq.Messages = append(openAiReq.Messages, gptMessageUser)

	jsonReq, err := json.Marshal(openAiReq)
	if err != nil {
		open.logger.Error(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonReq))
	if err != nil {
		open.logger.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		open.logger.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		open.logger.Error(err)
	}
	response := open.ParseGptResponse(body)
	return response
}

func (open *OpenaiStorage) ParseGptResponse(jsonResponse []byte) domain.OpenAiResponse {
	response := domain.OpenAiResponse{}
	err := json.Unmarshal(jsonResponse, &response)
	if err != nil {
		open.logger.Error(err)
	}
	return response
}

func (open *OpenaiStorage) SendToOpenAi(message *tgbotapi.Message, files []domain.File) domain.OpenAiResponse {
	fmt.Println("here")
	assistantID, err := open.GetLastAssistants()
	if err != nil {
		open.logger.Error(err)
	}
	fmt.Println("Assistant id: ", assistantID)

	open.FormAndSendImagesRequest(files)

	threadID, err := open.CreateThread(message, open.fileIds)
	if err != nil {
		open.logger.Error(err)
		return domain.OpenAiResponse{}
	}

	runID, err := open.RunThread(threadID, assistantID)
	if err != nil {
		open.logger.Error(err)
		return domain.OpenAiResponse{}
	}

	// Опрашиваем статус выполнения
	for {
		status, err := open.GetRunStatus(threadID, runID)
		if err != nil {
			open.logger.Error(err)
			return domain.OpenAiResponse{}
		}

		if status == "succeeded" {
			break
		} else if status == "failed" {
			open.logger.Error("Ошибка выполнения треда")
			return domain.OpenAiResponse{}
		}

		// Ждем перед следующим запросом
		time.Sleep(1 * time.Second)
	}

	// Получение ответа от ассистента
	messages, err := open.GetThreadMessages(threadID)
	if err != nil {
		open.logger.Error(err)
		return domain.OpenAiResponse{}
	}

	// Обработка и возврат ответа
	response := domain.OpenAiResponse{
		Choices: make([]domain.Choice, 0),
	}

	for i, msg := range messages {
		if msg.Role == "assistant" {
			response.Choices = append(response.Choices, domain.Choice{
				Index: i,
				Message: domain.GptAnswer{
					Role:    msg.Role,
					Content: msg.Content,
				},
				FinishReason: "stop",
			})
		}
	}

	return response
}

func (open *OpenaiStorage) SendImageToOpenAi(message *tgbotapi.Message) {
	//message.Photo[0].FileID
}

func (open *OpenaiStorage) CreateAssistant() (string, error) {
	assistantData := map[string]interface{}{
		"model":        "gpt-4o", // или другая модель
		"name":         "New assistent",
		"instructions": "Ты помощник в решении задач ТФЯ, отвечаешь очень понятно, развёрнуто, по-русски.",
		"tools": []map[string]interface{}{
			{
				"type": "code_interpreter",
			},
		},
	}

	jsonData, err := json.Marshal(assistantData)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/assistants", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ошибка при создании ассистента: %s", string(body))
	}

	var responseData struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return "", err
	}
	fmt.Println(responseData)
	return responseData.ID, nil
}
func (open *OpenaiStorage) CreateThread(message *tgbotapi.Message, fileIDs []string) (string, error) {
	// Формируем сообщение пользователя с вложениями
	gptMessageUser := domain.GptMessage{
		Role:        "user",
		Content:     message.Text,
		Attachments: []domain.Attachment{},
	}

	// Добавляем вложения
	for _, fileID := range fileIDs {
		attachment := domain.Attachment{
			FileID: fileID,
			Tools: []domain.Tool{
				{Type: "code_interpreter"},
				{Type: "file_search"},
			},
		}
		gptMessageUser.Attachments = append(gptMessageUser.Attachments, attachment)
	}

	messages := []domain.GptMessage{gptMessageUser}

	threadData := map[string]interface{}{
		"messages": messages,
	}

	jsonData, err := json.Marshal(threadData)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/threads", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ошибка при создании треда: %s", string(body))
	}

	var responseData struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return "", err
	}

	return responseData.ID, nil
}

func (open *OpenaiStorage) RunThread(threadID string, assistantID string) (string, error) {
	runData := map[string]interface{}{
		"assistant_id": assistantID,
	}

	jsonData, err := json.Marshal(runData)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.openai.com/v1/threads/%s/runs", threadID), bytes.NewBuffer(jsonData))
	fmt.Println("threadID: ", threadID)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ошибка при запуске треда: %s", string(body))
	}

	var responseData struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return "", err
	}

	// Вы можете добавить логику для опроса статуса выполнения, если это необходимо

	return responseData.ID, nil
}

func (open *OpenaiStorage) GetThreadMessages(threadID string) ([]domain.GptMessage, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages", threadID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+"skdDDujgqQ7Q3VeuYe6rbtrVjixAvT3BlbkFJSVYwSNKbNsm5NRoCIiiDflWapSsL06K4cGhe4-qSG21UTSFYIgF8GbxwuiePe9SZqnOuESWLAA")
	req.Header.Add("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ошибка при получении сообщений треда: %s", string(body))
	}

	var responseData struct {
		Messages []domain.GptMessage `json:"messages"`
	}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return nil, err
	}
	fmt.Println("here")
	fmt.Println(responseData)
	return responseData.Messages, nil
}

func (open *OpenaiStorage) Full() {
	assistantID, err := open.CreateAssistant()
	if err != nil {
		open.logger.Error(err)
	}

	open.assistantId = assistantID

}

func (open *OpenaiStorage) GetLastAssistants() (string, error) {
	req, err := http.NewRequest("GET", "https://api.openai.com/v1/assistants", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+"sk-rVjixAvT3BlbkFJSVYwSNKbNsm5NRoCIiiDflWapSsL06K4cGhe4-qSG21UTSFYIgF8GbxwuiePe9SZqnOuESWLAA")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ошибка при получении списка ассистентов: %s", string(body))
	}

	var responseData struct {
		LastID string `json:"last_id"`
	}

	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return "", err
	}

	fmt.Printf("Last ID: %s\n", responseData.LastID)
	return responseData.LastID, nil
}

func (open *OpenaiStorage) GetRunStatus(threadID string, runID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.openai.com/v1/threads/%s/runs/%s", threadID, runID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+"iiDflWapSsL06K4cGhe4-qSG21UTSFYIgF8GbxwuiePe9SZqnOuESWLAA")
	req.Header.Add("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ошибка при получении статуса выполнения: %s", string(body))
	}

	var responseData struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return "", err
	}

	return responseData.Status, nil
}
