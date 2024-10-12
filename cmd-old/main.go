package main

import (
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"tgbot/database"
	"tgbot/domain"
	tgDelivery "tgbot/internal/telegramAPI/delivery"
	userDelivery "tgbot/internal/user/delivery"
)

func main() {

	logger := CreateLogger()
	config := loadConfig(logger)
	db := database.ConnectToPostgreSQLDataBase(*logger)

	tgHandler := tgDelivery.NewTelegramHandler(logger)
	userHandler := userDelivery.NewUserHandler(logger, db)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Info("Запуск сервера на :8080")
		logger.Error(http.ListenAndServe(":8080", nil))
	}()

	tgHandler.CreateTelegramBot(config.Bot.TgKey)
	HandleTelegramMessages(tgHandler, userHandler, logger)

}

func loadConfig(logger *zap.SugaredLogger) domain.Config {
	file, err := os.Open("config.yml")
	logger.Info("Trying to open config.yml")
	if err != nil {
		logger.Error("load config failed: ", err)
	}
	defer file.Close()

	var config domain.Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		logger.Error("load config failed in decode: ", err)
	}
	return config
}

func CreateLogger() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()
	return sugar
}

func HandleTelegramMessages(telegramHandler *tgDelivery.TelegramHandler, userHandler *userDelivery.UserHandler, logger *zap.SugaredLogger) {
	bot := telegramHandler.TgBot.Bot

	updates, err := bot.GetUpdatesChan(telegramHandler.TgBot.Updates)
	if err != nil {
		logger.Error("Err by get updates chan: ", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		userExists, user := userHandler.CheckAccExists(update.Message.Chat.ID)
		if userExists == false {
			user = userHandler.RegisterUser(*update.Message)
		}

		logger.Info("[%s] %s", update.Message.From.UserName, update.Message.Text)
		user.State = userHandler.GetUserState(user.ChatID)
		switch user.State {
		case "start":
			handleStart(update.Message, userHandler)
		case "menu":
			handleMenu(bot, update.Message, userHandler)
		default:
			handleUnknown(bot, update.Message)
		}
	}
}

func handleStart(message *tgbotapi.Message, userHandler *userDelivery.UserHandler) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	msg.Text = ""
	userHandler.SetUserState(message.Chat.ID, "menu")
}

func handleMenu(bot *tgbotapi.BotAPI, message *tgbotapi.Message, userHandler *userDelivery.UserHandler) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	switch message.Text {
	case "/chat":
		msg.Text = "Тут будут чаты"
		bot.Send(msg)
	case "/manual":
		msg.Text = "тут будет readme"
		bot.Send(msg)
	case "/developers":
		msg.Text = "тут будут контакты разработчиков"
		bot.Send(msg)
	case "/profile":
		msg.Text = "тут будет профиль"
		bot.Send(msg)
	default:
		msg.Text = "Я не знаю, что сказать на это"
		bot.Send(msg)
	}
}

func handleUnknown(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Я не знаю, что сказать на это")
	bot.Send(msg)
}
