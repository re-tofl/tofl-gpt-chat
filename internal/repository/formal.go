package repository

import (
	"context"
	"fmt"
	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"github.com/re-tofl/tofl-gpt-chat/internal/utils"
	"net/http"
)

type FormalRepository struct {
	cfg *bootstrap.Config
}

func NewFormalRepository(cfg *bootstrap.Config) *FormalRepository {
	return &FormalRepository{
		cfg: cfg,
	}
}

func (r *FormalRepository) SendProblem(ctx context.Context, req domain.ParserResponse) (domain.FormalResponse, error) {
	resp, err := utils.SendRequestSugared(ctx, r.cfg.FormalURL+"/data", http.MethodPost, req)
	if err != nil {
		return domain.FormalResponse{}, fmt.Errorf("utils.SendRequestSugared: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return domain.FormalResponse{}, fmt.Errorf("resp.StatusCode from Formal: %v", resp.StatusCode)
	}

	var data domain.FormalResponse
	err = utils.DecodeBody(resp, &data)
	if err != nil {
		return domain.FormalResponse{}, fmt.Errorf("utils.DecodeBody: %w", err)
	}

	return data, nil
}
