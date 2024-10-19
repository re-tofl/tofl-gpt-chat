package usecase

import "github.com/prometheus/client_golang/prometheus"

type PrometheusMetrics struct {
	RegisteredUserCount prometheus.Gauge
	Hits                *prometheus.CounterVec
	Errors              *prometheus.CounterVec
	Methods             *prometheus.CounterVec
	RequestDuration     *prometheus.HistogramVec
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

	errorsInProject := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tg_chat_errors",
			Help: "Number of errors some type in bot.",
		}, []string{"error_type"},
	)

	methods := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tg_methods",
			Help: "called methods.",
		}, []string{"method"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_request_duration_seconds",
			Help:    "Histogram of request to llm durations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	prometheus.MustRegister(activeSessionsCount, hits, errorsInProject, methods, requestDuration)

	return &PrometheusMetrics{
		RegisteredUserCount: activeSessionsCount,
		Hits:                hits,
		Errors:              errorsInProject,
		Methods:             methods,
		RequestDuration:     requestDuration,
	}
}
