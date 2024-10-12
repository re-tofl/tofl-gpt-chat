package app

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/re-tofl/tofl-gpt-chat/internal/tgchat/bootstrap"
	"net/http"
)

type PollEntrypoint struct {
	Config *bootstrap.Config
	server *http.Server
}

func (e *PollEntrypoint) Init(ctx context.Context) error {
	mux := http.DefaultServeMux
	mux.Handle("/metrics", promhttp.Handler())

	e.server = &http.Server{
		Handler: mux,
		Addr:    "0.0.0.0:" + e.Config.ServerPort,
	}

	return nil
}

func (e *PollEntrypoint) Run(ctx context.Context) error {
	return e.server.ListenAndServe()
}

func (e *PollEntrypoint) Close() error {
	return e.server.Close()
}
