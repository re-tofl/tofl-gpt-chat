package usecase

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"go.uber.org/zap"
	"strconv"
	"sync"
)

type RatingRepository interface {
	SaveRating(ctx context.Context, rating domain.Rating) error
}

type LLMRepository interface {
	SendProblem(ctx context.Context, req domain.LLMRequest) (domain.LLMProblemResponse, error)
	SendTheory(ctx context.Context, req domain.LLMRequest) (domain.LLMTheoryResponse, error)
}

type ParserRepository interface {
	SendProblem(ctx context.Context, req domain.LLMProblemResponse) (domain.UnionParserResponse, error)
}

type FormalRepository interface {
	SendProblem(ctx context.Context, req domain.ParserResponse) (domain.FormalResponse, error)
}

type TaskUsecase struct {
	log            *zap.SugaredLogger
	metrics        *PrometheusMetrics
	llm            LLMRepository
	parser         ParserRepository
	formal         FormalRepository
	ratingRepo     RatingRepository
	mu             sync.Mutex
	userContextIDs map[int64]int
}

func NewTaskUsecase(log *zap.SugaredLogger, llm LLMRepository, parser ParserRepository,
	formal FormalRepository, ratingRepo RatingRepository) *TaskUsecase {
	return &TaskUsecase{
		log:            log,
		metrics:        NewPrometheusMetrics(),
		llm:            llm,
		parser:         parser,
		formal:         formal,
		ratingRepo:     ratingRepo,
		userContextIDs: make(map[int64]int),
	}
}

func (t *TaskUsecase) SolveProblem(ctx context.Context, message domain.Message) (domain.UnionProblemResponse, error) {
	parserReq, err := t.llm.SendProblem(ctx, domain.LLMRequest{Type: 1, Prompt: message.Text})
	if err != nil {
		return domain.UnionProblemResponse{}, err
	}
	t.log.Infow("problem from llm", "trs", parserReq.TRS,
		"interpretation", parserReq.Interpretation)

	formalReq, err := t.parser.SendProblem(ctx, parserReq)
	if errors.Is(err, domain.ErrBadRequest) {
		return domain.UnionProblemResponse{Error: formalReq.Error}, err
	}
	if err != nil {
		return domain.UnionProblemResponse{}, err
	}
	t.log.Info("problem parsed successfully")

	resp, err := t.formal.SendProblem(ctx, formalReq.Success)
	if err != nil {
		return domain.UnionProblemResponse{}, err
	}
	t.log.Info("problem formalized successfully")

	return domain.UnionProblemResponse{Success: resp}, nil
}

func (t *TaskUsecase) AnswerTheory(ctx context.Context, message domain.Message) (domain.LLMTheoryResponse, error) {
	//h.handleGptTextMessage(ctx, message)
	res, err := t.llm.SendTheory(ctx, domain.LLMRequest{Type: 0, Prompt: message.Text})
	if err != nil {
		return domain.LLMTheoryResponse{}, err
	}

	t.mu.Lock()
	t.userContextIDs[message.ChatID] = res.ContextID
	t.mu.Unlock()

	return res, nil
}

func (t *TaskUsecase) RateTheory(ctx context.Context, message domain.Message) error {
	t.mu.Lock()
	contextID := t.userContextIDs[message.ChatID]
	t.mu.Unlock()

	ratingValue, err := strconv.ParseFloat(message.Text, 64)
	if err != nil || ratingValue < 1 || ratingValue > 10 {
		return fmt.Errorf("неверное значение: %s. Должно быть от 1 до 10", message.Text)
	}

	t.metrics.ResponseRating.WithLabelValues(message.UserName).Observe(ratingValue)
	return t.ratingRepo.SaveRating(ctx, domain.Rating{
		ChatID:    message.ChatID,
		ContextID: contextID,
		Rating:    int(ratingValue),
	})
}
