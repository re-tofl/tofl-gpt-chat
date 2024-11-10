package telegram

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	task2 "github.com/re-tofl/tofl-gpt-chat/internal/delivery/openai"
	"github.com/re-tofl/tofl-gpt-chat/internal/delivery/task"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
)

type Handler struct {
	cfg         *bootstrap.Config
	log         *zap.SugaredLogger
	bot         *tgbotapi.BotAPI
	taskHandler *task.THandler
	openHandler *task2.OpenHandler
}

func NewHandler(cfg *bootstrap.Config, log *zap.SugaredLogger, t *task.THandler, o *task2.OpenHandler) *Handler {
	return &Handler{
		cfg:         cfg,
		log:         log,
		taskHandler: t,
		openHandler: o,
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
		h.log.Infow("received photo or file",
			"username", update.Message.From.UserName)

		h.handleProblemWithImages(ctx, update.Message)
		return
	}

	if update.Message.Document != nil {
		h.log.Infow("received doc",
			"username", update.Message.From.UserName)

		h.handleProblemWithImages(ctx, update.Message)
		return
	}

	if update.Message.IsCommand() {
		h.processCommand(ctx, update.Message)
	}
}

func (h *Handler) handleProblemWithImages(ctx context.Context, message *tgbotapi.Message) {
	files := h.SaveMedia(message)
	h.openHandler.SendToOpenAi(ctx, message, files)
}
func (h *Handler) SaveMedia(message *tgbotapi.Message) []domain.File {
	files := make([]domain.File, 0)

	// Проверяем, есть ли фото
	if message.Photo != nil && len(*message.Photo) > 0 {
		images := *message.Photo
		fileID := images[len(images)-1].FileID // берем изображение наилучшего качества

		file, err := h.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
		if err != nil {
			log.Printf("Ошибка при получении файла: %v\n", err)
			return files
		}

		fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", h.bot.Token, file.FilePath)

		fmt.Println("URL для изображения:", fileURL)

		resp, err := http.Get(fileURL)
		if err != nil {
			log.Printf("Ошибка при загрузке изображения: %v\n", err)
			return files
		}
		defer resp.Body.Close()

		fileName := fmt.Sprintf("image_%s.jpg", file.FileID)
		filePath := filepath.Join("upload", fileName)

		outFile, err := os.Create(filePath)
		if err != nil {
			log.Printf("Ошибка при создании файла: %v\n", err)
			return files
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			log.Printf("Ошибка при сохранении файла: %v\n", err)
			return files
		}

		fmt.Println("Изображение сохранено в:", filePath)
		files = append(files, domain.File{Name: fileName, Path: filePath})
	}

	// Проверяем, есть ли документ
	if message.Document != nil {
		fileID := message.Document.FileID

		file, err := h.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
		if err != nil {
			log.Printf("Ошибка при получении документа: %v\n", err)
			return files
		}

		fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", h.bot.Token, file.FilePath)

		fmt.Println("URL для документа:", fileURL)

		resp, err := http.Get(fileURL)
		if err != nil {
			log.Printf("Ошибка при загрузке документа: %v\n", err)
			return files
		}
		defer resp.Body.Close()

		fileExtension := filepath.Ext(message.Document.FileName)
		if fileExtension == "" {
			// Пытаемся определить расширение из пути файла
			fileExtension = filepath.Ext(file.FilePath)
		}

		if fileExtension == "" {
			// Если расширение не найдено, добавляем стандартное
			fileExtension = ".dat"
		}

		fileName := fmt.Sprintf("%s_%d%s", message.Document.FileName, time.Now().UnixNano(), fileExtension)
		filePath := filepath.Join("upload", fileName)

		outFile, err := os.Create(filePath)
		if err != nil {
			log.Printf("Ошибка при создании файла: %v\n", err)
			return files
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			log.Printf("Ошибка при сохранении файла: %v\n", err)
			return files
		}

		fmt.Println("Документ сохранен в:", filePath)
		files = append(files, domain.File{Name: fileName, Path: filePath})
	}

	return files
}

func (h *Handler) processCommand(ctx context.Context, message *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(message.Chat.ID, "")

	switch message.Command() {
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
