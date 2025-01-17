package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"github.com/re-tofl/tofl-gpt-chat/internal/utils"
	"net/http"
	"sync"
)

type LLMRepository struct {
	sync.Once
	mu  sync.Mutex
	cfg *bootstrap.Config
}

func NewLLMRepository(cfg *bootstrap.Config) *LLMRepository {
	return &LLMRepository{
		cfg: cfg,
	}
}

func (r *LLMRepository) sendRequest(ctx context.Context, url string, method string, req domain.LLMRequest) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	resp, err := utils.SendRequestSugared(ctx, url, method, req)
	return resp, err
}

func (r *LLMRepository) GetClosestQuestions(ctx context.Context, req domain.LLMRequest) (domain.LLMClosestQuestionsResponse, error) {
	resp, err := r.sendRequest(ctx, r.cfg.LLMURL+"/theory-closest-questions", "POST", req)
	if err != nil {
		return domain.LLMClosestQuestionsResponse{}, fmt.Errorf("utils.SendRequestSugared: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return domain.LLMClosestQuestionsResponse{}, fmt.Errorf("resp.StatusCode from LLM: %v", resp.StatusCode)
	}

	var data domain.LLMClosestQuestionsResponse
	err = utils.DecodeBody(resp, &data)
	if err != nil {
		return domain.LLMClosestQuestionsResponse{}, fmt.Errorf("utils.DecodeBody: %w", err)
	}

	return data, nil
}

func (r *LLMRepository) SendTheory(ctx context.Context, req domain.LLMRequest) (domain.LLMTheoryResponse, error) {
	resp, err := r.sendRequest(ctx, r.cfg.LLMURL+"/process", "POST", req)
	if err != nil {
		return domain.LLMTheoryResponse{}, fmt.Errorf("utils.SendRequestSugared: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return domain.LLMTheoryResponse{}, fmt.Errorf("resp.StatusCode from LLM: %v", resp.StatusCode)
	}

	var data domain.LLMTheoryResponse
	err = utils.DecodeBody(resp, &data)
	if err != nil {
		return domain.LLMTheoryResponse{}, fmt.Errorf("utils.DecodeBody: %w", err)
	}

	return data, nil
}

func (r *LLMRepository) SendProblem(ctx context.Context, req domain.LLMRequest) (domain.LLMProblemResponse, error) {
	resp, err := r.sendRequest(ctx, r.cfg.LLMURL+"/process", "POST", req)
	if err != nil {
		return domain.LLMProblemResponse{}, fmt.Errorf("utils.SendRequestSugared: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return domain.LLMProblemResponse{}, fmt.Errorf("resp.StatusCode from LLM: %v", resp.StatusCode)
	}

	var parserReq domain.LLMProblemResponse
	err = json.NewDecoder(resp.Body).Decode(&parserReq)
	if err != nil {
		return domain.LLMProblemResponse{}, fmt.Errorf("json.NewDecoder: %w", err)
	}

	return parserReq, nil
}
