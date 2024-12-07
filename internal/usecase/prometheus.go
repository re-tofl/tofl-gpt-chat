package usecase

import "github.com/prometheus/client_golang/prometheus"

type PrometheusMetrics struct {
	RegisteredUserCount prometheus.Gauge
	Hits                *prometheus.CounterVec
	Errors              prometheus.Counter
	ResponseRating      *prometheus.HistogramVec
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

	responseRating := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_rating",
			Help:    "Histogram of user ratings for the response (1 to 10).",
			Buckets: prometheus.LinearBuckets(1, 1, 10),
		},
		[]string{"username"},
	)

	prometheus.MustRegister(activeSessionsCount, hits, errorsInProject, responseRating)

	return &PrometheusMetrics{
		RegisteredUserCount: activeSessionsCount,
		Hits:                hits,
		Errors:              errorsInProject,
		ResponseRating:      responseRating,
	}
}
