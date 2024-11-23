package telegram

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"go.uber.org/zap"
)

type OpenAiUsecase interface {
	SaveMedia(ctx context.Context, message *tgbotapi.Message, bot *tgbotapi.BotAPI) string
	SendToGpt(ctx context.Context, message *tgbotapi.Message) string
}

type SpeechUsecase interface {
	ConvertSpeechToText(ctx context.Context, filePath string) string
}

type Handler struct {
	cfg        *bootstrap.Config
	log        *zap.SugaredLogger
	bot        *tgbotapi.BotAPI
	openAiUC   OpenAiUsecase
	speechUC   SpeechUsecase
	mu         sync.Mutex
	userStates map[int64]string
}

func NewHandler(cfg *bootstrap.Config, log *zap.SugaredLogger,
	o OpenAiUsecase, s SpeechUsecase) *Handler {
	return &Handler{
		cfg:      cfg,
		log:      log,
		openAiUC: o,
		speechUC: s,
	}
}

func (h *Handler) Listen(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(h.cfg.TGBotToken)
	if err != nil {
		h.log.Errorw("tgbotapi.NewBotAPI", zap.Error(err))
		return fmt.Errorf("tgbotapi.NewBotAPI: %w", err)
	}

	h.bot = bot
	h.log.Infof("logged in TG with username %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		h.log.Errorw("bot.GetUpdatesChan", zap.Error(err))
		return fmt.Errorf("bot.GetUpdatesChan: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			bot.StopReceivingUpdates()
			return nil
		case update := <-updates:
			// TODO: сделать многопоточным
			h.processUpdate(ctx, &update)
		}
	}
}

func (h *Handler) Send(c tgbotapi.Chattable) *tgbotapi.Message {
	send, err := h.bot.Send(c)
	if err != nil {
		h.log.Errorw("bot.Send", zap.Error(err))
		return nil
	}

	return &send
}

func (h *Handler) processUpdate(ctx context.Context, update *tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	messageText := update.Message.Text
	if messageText == "" {
		messageText = update.Message.Caption
		update.Message.Text = messageText
	}

	h.log.Infow("received message",
		"username", update.Message.From.UserName,
		"message", messageText)

	if update.Message.Photo != nil || update.Message.Document != nil {
		h.log.Infow("received photo or file from ",
			"username", update.Message.From.UserName)

		h.handleProblemWithImages(ctx, update.Message)
		return
	}

	if update.Message.Voice != nil {
		h.handleVoice(ctx, update.Message, h.bot)
	}

	fmt.Println(update.Message)

	if update.Message.IsCommand() {
		h.processCommand(ctx, update.Message)
	} else {
		h.handleMessage(ctx, update.Message)
	}

}

func (h *Handler) handleProblemWithImages(ctx context.Context, message *tgbotapi.Message) {
	gptFileResponse := h.openAiUC.SaveMedia(ctx, message, h.bot)
	reply := tgbotapi.NewMessage(message.Chat.ID, gptFileResponse)
	h.Send(reply)
}

func (h *Handler) handleVoice(ctx context.Context, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	fmt.Println("here")
	fileId := message.Voice.FileID
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileId})
	if err != nil {
		h.log.Error(err)
	}

	filePath, err := h.SaveAndDownloadVoice(file.FilePath, file.FileID)
	textFromSpeech := h.speechUC.ConvertSpeechToText(ctx, filePath)

	reply := tgbotapi.NewMessage(message.Chat.ID, textFromSpeech)
	h.Send(reply)
}

func (h *Handler) SaveAndDownloadVoice(tgFilePath string, fileName string) (string, error) {
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", h.cfg.TGBotToken, tgFilePath)

	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("ошибка при загрузке файла: %w", err)
	}
	defer resp.Body.Close()

	filePath := filepath.Join("upload/voices", fileName+".ogg")
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
func (h *Handler) handleGptTextMessage(ctx context.Context, message *tgbotapi.Message) {
	gptResponse := h.openAiUC.SendToGpt(ctx, message)
	reply := tgbotapi.NewMessage(message.Chat.ID, gptResponse)
	h.Send(reply)
}

func (h *Handler) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(message.Chat.ID, "")

	if state, exists := userStates[message.Chat.ID]; exists {
		switch state {
		case "awaiting_gpt_question":
			question := message.Text
			reply.Text = fmt.Sprintf("Ты задал вопрос: %s. Отправляю в GPT...", question)
			h.Send(reply)
			h.handleGptTextMessage(ctx, message)

			delete(userStates, message.Chat.ID)
		default:
			reply.Text = "Неизвестное состояние."
		}
		return
	}

	h.processCommand(ctx, message)
}

func (h *Handler) processCommand(ctx context.Context, message *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(message.Chat.ID, "")

	switch message.Command() {
	case "gpt":
		reply.Text = "Введи вопрос к gpt:"
		userStates[message.Chat.ID] = "awaiting_gpt_question"
		h.Send(reply)
	case "theory":
		reply.Text = "Ответ на теорию"
		h.Send(reply)
	case "problem":
		reply.Text = "Ответ на задачу"
		h.Send(reply)
	case "imageProblem":
		reply.Text = "Ответ на задачу с фотографиями"
		h.handleProblemWithImages(ctx, message)
		h.Send(reply)
	case "developers":
		reply.Text = "тут будут контакты разработчиков"
		h.Send(reply)
	default:
		reply.Text = "Неизвестная команда"
		h.Send(reply)
	}
}

var userStates = make(map[int64]string)
