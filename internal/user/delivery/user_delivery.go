package delivery

import (
	"database/sql"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"tgbot/domain"
	userRepo "tgbot/internal/user/repository"
	userUsecase "tgbot/internal/user/usecase"
)

type UserHandler struct {
	Logger            *zap.SugaredLogger
	users             *userRepo.UserStorage
	prometheusMetrics *PrometheusMetrics
}

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

func NewUserHandler(Logger *zap.SugaredLogger, db *sql.DB) *UserHandler {
	return &UserHandler{
		Logger:            Logger,
		users:             userRepo.NewUserStorage(Logger, db),
		prometheusMetrics: NewPrometheusMetrics(),
	}
}

func (u *UserHandler) SetUserState(chatID int64, state string) {
	userUsecase.SetUserState(chatID, state, u.users)
	u.prometheusMetrics.Methods.WithLabelValues("SwitchState").Inc()
}

func (u *UserHandler) GetUserState(chatID int64) string {
	userState := userUsecase.GetUserState(chatID, u.users)
	u.prometheusMetrics.Methods.WithLabelValues("GetState").Inc()
	return userState
}

func (u *UserHandler) CheckAccExists(chatID int64) (bool, domain.User) {
	u.prometheusMetrics.Methods.WithLabelValues("CheckAcc").Inc()
	exists, user := userUsecase.CheckAccExists(chatID, u.users)
	return exists, user
}

func (u *UserHandler) RegisterUser(message tgbotapi.Message) domain.User {
	user := domain.User{}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.ChatID = message.Chat.ID
	user.State = "start"
	user.Nickname = message.Chat.UserName

	user.Id = userUsecase.Register(&user, u.users)
	u.prometheusMetrics.RegisteredUserCount.Inc()
	return user
}
