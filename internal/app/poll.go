package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/delivery/telegram"
	"github.com/re-tofl/tofl-gpt-chat/internal/depgraph"
)

type PollEntrypoint struct {
	Config *bootstrap.Config
	server *http.Server
	tgbot  *telegram.Handler
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
		Addr:    "0.0.0.0:" + e.Config.ServerPort,
	}

	e.tgbot = telegram.NewHandler(e.Config, logger)

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
