package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
)

type Handler struct {
	cfg *bootstrap.Config
	log *zap.SugaredLogger
	bot *tgbotapi.BotAPI
}

func NewHandler(cfg *bootstrap.Config, log *zap.SugaredLogger) *Handler {
	return &Handler{
		cfg: cfg,
		log: log,
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

	h.log.Infow("received message",
		"username", update.Message.From.UserName,
		"message", update.Message.Text)

	if update.Message.IsCommand() {
		h.processCommand(ctx, update.Message)
	}
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
	case "developers":
		reply.Text = "тут будут контакты разработчиков"
		h.Send(reply)
	default:
		reply.Text = "Неизвестная команда"
		h.Send(reply)
	}
}
