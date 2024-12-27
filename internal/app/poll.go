package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/re-tofl/tofl-gpt-chat/internal/adapters"
	"golang.org/x/sync/errgroup"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/delivery/telegram"
	"github.com/re-tofl/tofl-gpt-chat/internal/depgraph"
	"github.com/re-tofl/tofl-gpt-chat/internal/repository"
	"github.com/re-tofl/tofl-gpt-chat/internal/usecase"
)

type PollEntrypoint struct {
	Config *bootstrap.Config
	server *http.Server
	tgbot  *telegram.Handler
	fw     *adapters.FileWriter
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

	openAiRepo := repository.NewOpenaiStorage(logger, e.Config)
	speechRepo := repository.NewSpeechStorage(logger, e.Config)

	fw, err := dg.GetFileWriter("rating.json")
	if err != nil {
		return fmt.Errorf("dg.GetFileWriter: %w", err)
	}

	e.fw = fw
	ratingRepo := repository.NewRatingRepository(fw)

	speechUC := usecase.NewSpeechUsecase(speechRepo)
	openAiUC := usecase.NewOpenAiUseCase(openAiRepo)
	taskUC := usecase.NewTaskUsecase(ratingRepo)

	mongoAdapter := adapters.NewAdapterMongo(e.Config)
	//postgresAdapter := adapters.NewAdapterPG(e.Config)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = mongoAdapter.Init(ctx); err != nil {
		log.Fatalf("Не удалось инициализировать MongoAdapter: %v", err)
	}

	//if err = postgresAdapter.Init(ctx); err != nil {
	//	log.Fatalf("Не удалось инициализировать PostgresAdapter: %v", err)
	//}

	searchRepo := repository.NewSearchStorage(logger)
	searchUC := usecase.NewSearchUseCase(searchRepo)

	e.tgbot = telegram.NewHandler(e.Config, logger, openAiUC, speechUC, taskUC, mongoAdapter, searchUC)

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
	err := e.server.Close()
	if err != nil {
		return fmt.Errorf("e.server.Close: %w", err)
	}
	err = e.fw.Close()
	if err != nil {
		return fmt.Errorf("e.fw.Close: %w", err)
	}
	return nil
}
