package repository

import (
	"context"
	"fmt"
	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"github.com/re-tofl/tofl-gpt-chat/internal/utils"
	"net/http"
)

type ParserRepository struct {
	cfg *bootstrap.Config
}

func NewParserRepository(cfg *bootstrap.Config) *ParserRepository {
	return &ParserRepository{
		cfg: cfg,
	}
}

func (r *ParserRepository) SendProblem(ctx context.Context, req domain.LLMProblemResponse) (domain.UnionParserResponse, error) {
	resp, err := utils.SendRequestSugared(ctx, r.cfg.ParserURL+"/parse", "POST", req)
	if err != nil {
		return domain.UnionParserResponse{}, fmt.Errorf("utils.SendRequestSugared: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusBadRequest {
			return domain.UnionParserResponse{}, fmt.Errorf("resp.StatusCode from Parser: %v", resp.StatusCode)
		}

		var data domain.ParserErrorResponse
		err = utils.DecodeBody(resp, &data)
		if err != nil {
			return domain.UnionParserResponse{}, fmt.Errorf("utils.DecodeBody: %w", err)
		}

		return domain.UnionParserResponse{Error: data}, fmt.Errorf("%w Parser", domain.ErrBadRequest)
	}

	var data domain.ParserResponse
	err = utils.DecodeBody(resp, &data)
	if err != nil {
		return domain.UnionParserResponse{}, fmt.Errorf("utils.DecodeBody: %w", err)
	}

	return domain.UnionParserResponse{Success: data}, nil
}
