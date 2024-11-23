package usecase

import "github.com/prometheus/client_golang/prometheus"

type PrometheusMetrics struct {
	RegisteredUserCount prometheus.Gauge
	Hits                *prometheus.CounterVec
	Errors              prometheus.Counter
	GoodResponsesLLM    prometheus.Counter
	BadResponsesLLM     prometheus.Counter
}

func NewPrometheusMetrics() *PrometheusMetrics {
	activeSessionsCount := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "registered_user_total",
			Help: "Total number of users.",
		},
	)

	hits := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "menu_hits",
			Help: "Total number of hits in menu.",
		}, []string{"status", "path"},
	)

	errorsInProject := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "tg_chat_errors",
			Help: "Number of errors some type in bot.",
		},
	)

	goodResponsesLLM := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "llm_responses",
			Help: "Number of good responses from LLM.",
		},
	)

	badResponsesLLM := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "llm_errors",
			Help: "Number of bad responses from LLM.",
		},
	)

	prometheus.MustRegister(activeSessionsCount, hits, errorsInProject, goodResponsesLLM, badResponsesLLM)

	return &PrometheusMetrics{
		RegisteredUserCount: activeSessionsCount,
		Hits:                hits,
		Errors:              errorsInProject,
		GoodResponsesLLM:    goodResponsesLLM,
		BadResponsesLLM:     badResponsesLLM,
	}
}
