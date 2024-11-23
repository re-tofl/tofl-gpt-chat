package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"go.uber.org/zap"
	"net/http"
)

// TODO: тут я уже решил забить на чистую архитектуру, потому что под нее надо рефакторить весь проект

func (h *Handler) requestParser(ctx context.Context, message *tgbotapi.Message) (*http.Response, error) {
	jsonStr := message.Text
	req, err := http.NewRequestWithContext(ctx, "POST", h.cfg.ParserURL+"/parse", bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		h.log.Errorw("http.NewRequest", zap.Error(err))
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.log.Errorw("client.Do", zap.Error(err))
		h.Send(tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка"))
		return nil, fmt.Errorf("client.Do: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.log.Errorw("resp.Body.Close", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		h.log.Errorw("resp.StatusCode", zap.Error(err))
		h.Send(tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка"))
		return nil, fmt.Errorf("resp.StatusCode: %w", err)
	}

	return resp, nil
}

func (h *Handler) requestFormal(ctx context.Context, message *tgbotapi.Message, resp *http.Response) error {
	req, err := http.NewRequestWithContext(ctx, "POST", h.cfg.FormalURL+"/data", resp.Body)
	if err != nil {
		h.log.Errorw("http.NewRequest", zap.Error(err))
		return fmt.Errorf("http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		h.log.Errorw("client.Do", zap.Error(err))
		return fmt.Errorf("client.Do: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.log.Errorw("resp.Body.Close", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		h.log.Errorw("resp.StatusCode", zap.Error(err))
		return fmt.Errorf("resp.StatusCode: %w", err)
	}

	var data domain.FormalResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		h.log.Errorw("json.NewDecoder", zap.Error(err))
		return fmt.Errorf("json.NewDecoder: %w", err)
	}

	var msg string
	for _, result := range data.Result {
		msg += result.Data + "\n"
	}

	h.Send(tgbotapi.NewMessage(message.Chat.ID, "The problem has been solved:\n"+msg))

	return nil
}

func (h *Handler) Problem(ctx context.Context, message *tgbotapi.Message) error {
	resp, err := h.requestParser(ctx, message)
	if err != nil {
		return err
	}

	err = h.requestFormal(ctx, message, resp)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) Theory(ctx context.Context) error {
	panic("implement me")
}
