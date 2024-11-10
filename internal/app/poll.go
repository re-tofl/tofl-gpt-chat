package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	task2 "github.com/re-tofl/tofl-gpt-chat/internal/delivery/openai"
	"github.com/re-tofl/tofl-gpt-chat/internal/delivery/task"
	"github.com/re-tofl/tofl-gpt-chat/internal/usecase"
	"golang.org/x/sync/errgroup"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/delivery/telegram"
	"github.com/re-tofl/tofl-gpt-chat/internal/depgraph"
	userRep "github.com/re-tofl/tofl-gpt-chat/internal/repository"
)

type PollEntrypoint struct {
	Config *bootstrap.Config
	server *http.Server
	tgbot  *telegram.Handler
	task   *task.THandler
}

func (e *PollEntrypoint) Init(ctx context.Context) error {
	dg := depgraph.NewDepGraph()

	logger, err := dg.GetLogger()
	if err != nil {
		return fmt.Errorf("dg.GetLogger: %w", err)
	}

	mux := http.DefaultServeMux
	mux.Handle("/metrics", promhttp.Handler())

	e.server = &http.Server{
		Handler: mux,
		Addr:    "127.0.0.1:" + e.Config.ServerPort,
	}

	taskRepo := userRep.NewTaskStorage(nil, nil, logger, e.Config)
	openAiRepo := userRep.NewOpenaiStorage(logger, e.Config)

	openHandler := task2.NewOpenHandler(e.Config, logger)
	e.task = task.NewTaskHandler(e.Config, logger, taskRepo, openAiRepo)

	usecase.NewUserHandler(logger, userRep.NewUserStorage(logger, nil))

	e.tgbot = telegram.NewHandler(e.Config, logger, e.task, openHandler)

	return nil
}

func (e *PollEntrypoint) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(e.server.ListenAndServe)
	eg.Go(func() error {
		return e.tgbot.Listen(ctx)
	})

	return eg.Wait()
}

func (e *PollEntrypoint) Close() error {
	return e.server.Close()
}
