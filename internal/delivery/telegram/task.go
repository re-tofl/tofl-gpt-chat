package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

// TODO: тут я уже решил забить на чистую архитектуру, потому что под нее надо рефакторить весь проект

func (h *Handler) requestParser(ctx context.Context, message *tgbotapi.Message) (*domain.ParserResponse, error) {
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
		h.log.Errorw("resp.StatusCode", "statusCode", resp.StatusCode, zap.Error(err))
		if resp.StatusCode != http.StatusBadRequest {
			return nil, fmt.Errorf("resp.StatusCode Parser: %v", resp.StatusCode)
		}

		var data domain.ParserErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			h.log.Errorw("json.Unmarshal", zap.Error(err))
			h.Send(tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка"))
			return nil, fmt.Errorf("json.Unmarshal: %w", err)
		}

		if data.ErrorTrs != nil {
			h.log.Errorw("data.ErrorTrs", zap.Error(err))
			h.Send(tgbotapi.NewMessage(message.Chat.ID, "Ошибки TRS"))
			for _, err := range data.ErrorTrs {
				h.Send(tgbotapi.NewMessage(message.Chat.ID, err))
			}
		}

		if data.ErrorInterpretation != nil {
			h.log.Errorw("data.ErrorInterpretation", zap.Error(err))
			h.Send(tgbotapi.NewMessage(message.Chat.ID, "Ошибки интерпретации"))
			for _, err := range data.ErrorInterpretation {
				h.Send(tgbotapi.NewMessage(message.Chat.ID, err))
			}
		}

		return nil, fmt.Errorf("resp.StatusCode Parser: %v", resp.StatusCode)
	}

	var parserResp domain.ParserResponse
	err = json.NewDecoder(resp.Body).Decode(&parserResp)
	if err != nil {
		h.log.Errorw("json.Unmarshal", zap.Error(err))
		h.Send(tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка"))
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return &parserResp, nil
}

func (h *Handler) requestFormal(ctx context.Context, message *tgbotapi.Message, parserResponse *domain.ParserResponse) error {
	b, err := json.Marshal(parserResponse)
	if err != nil {
		h.log.Errorw("json.Marshal", zap.Error(err))
		return fmt.Errorf("json.Marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.cfg.FormalURL+"/data", bytes.NewBuffer(b))
	if err != nil {
		h.log.Errorw("http.NewRequest", zap.Error(err))
		return fmt.Errorf("http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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
		h.log.Errorw("resp.StatusCode", "statusCode", resp.StatusCode, zap.Error(err))
		return fmt.Errorf("resp.StatusCode from Formal %v", resp.StatusCode)
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
	parserReq, err := h.requestProblemLLM(ctx, message)
	if err != nil {
		return err
	}
	h.log.Info("get from llm: interpret: \n", parserReq.Interpretation, "trs: \n", parserReq.TRS)

	message.Text = parserReq.TRS + parserReq.Interpretation

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

func (h *Handler) requestTheoryLLM(ctx context.Context, message *tgbotapi.Message) error {
	var data domain.LLMRequest
	data.Prompt = message.Text
	data.Type = 0

	b, err := json.Marshal(data)
	if err != nil {
		h.log.Errorw("json.Marshal", zap.Error(err))
		return fmt.Errorf("json.Marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.cfg.LLMURL+"/process", bytes.NewBuffer(b))
	if err != nil {
		h.log.Errorw("http.NewRequest", zap.Error(err))
		return fmt.Errorf("http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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
		h.log.Errorw("resp.StatusCode", "statusCode", resp.StatusCode, zap.Error(err))
		return fmt.Errorf("resp.StatusCode from LLM %v", resp.StatusCode)
	}

	var dataLLM domain.LLMTheoryResponse
	err = json.NewDecoder(resp.Body).Decode(&dataLLM)
	fmt.Println(resp.Body)

	if err != nil {
		h.log.Errorw("json.NewDecoder", zap.Error(err))
		return fmt.Errorf("json.NewDecoder: %w", err)
	}

	h.mu.Lock()
	h.userContextIDs[message.Chat.ID] = dataLLM.ContextID
	h.mu.Unlock()

	h.Send(tgbotapi.NewMessage(message.Chat.ID, dataLLM.Response))
	return nil
}

func (h *Handler) requestProblemLLM(ctx context.Context, message *tgbotapi.Message) (domain.ParserRequest, error) {
	var data domain.LLMRequest
	data.Prompt = message.Text
	data.Type = 1

	b, err := json.Marshal(data)
	if err != nil {
		h.log.Errorw("json.Marshal", zap.Error(err))
		return domain.ParserRequest{}, fmt.Errorf("json.Marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.cfg.LLMURL+"/process", bytes.NewBuffer(b))
	if err != nil {
		h.log.Errorw("http.NewRequest", zap.Error(err))
		return domain.ParserRequest{}, fmt.Errorf("http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.log.Errorw("client.Do", zap.Error(err))
		return domain.ParserRequest{}, fmt.Errorf("client.Do: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.log.Errorw("resp.Body.Close", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		h.log.Errorw("resp.StatusCode", "statusCode", resp.StatusCode, zap.Error(err))
		return domain.ParserRequest{}, fmt.Errorf("resp.StatusCode from LLM %v", resp.StatusCode)
	}

	var parserReq domain.ParserRequest
	err = json.NewDecoder(resp.Body).Decode(&parserReq)
	fmt.Println(resp.Body)

	if err != nil {
		h.log.Errorw("json.NewDecoder", zap.Error(err))
		return domain.ParserRequest{}, fmt.Errorf("json.NewDecoder: %w", err)
	}

	return parserReq, nil
}

func (h *Handler) Theory(ctx context.Context, message *tgbotapi.Message) error {
	//h.handleGptTextMessage(ctx, message)
	err := h.requestTheoryLLM(ctx, message)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) RateTheory(ctx context.Context, message *tgbotapi.Message) error {
	contextID := h.userContextIDs[message.Chat.ID]
	return h.taskUC.RateTheory(ctx, message, contextID)
}
