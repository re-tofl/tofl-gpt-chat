package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"tgbot/database"
	"tgbot/domain"
	tgDelivery "tgbot/internal/telegramAPI/delivery"
	userDelivery "tgbot/internal/user/delivery"
)

func main() {
	logger := CreateLogger()
	db := database.ConnectToPostgreSQLDataBase(*logger)

	tgHandler := tgDelivery.NewTelegramHandler(logger)
	userHandler := userDelivery.NewUserHandler(logger, db)

	config := loadConfig(logger)
	fmt.Println("tgkey: ", config.Bot.TgKey)
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
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		userExists, user := userHandler.CheckAccExists(update.Message.Chat.ID)
		if userExists == false {
			user = userHandler.RegisterUser(*update.Message)
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		user.State = userHandler.GetUserState(user.ChatID)
		switch user.State {
		case "start":
			handleStart(bot, update.Message, userHandler, logger)
			/*
				case "menu":
					handleMenu(bot, update.Message, &handlers)
				default:
					handleUnknown(bot, update.Message, &handlers)*/
		}
	}
}

func handleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message, userHandler *userDelivery.UserHandler, logger *zap.SugaredLogger) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	msg.Text = "Добро пожаловать! Это бот-помощник по ТФЯ."
	userHandler.SetUserState(message.Chat.ID, "menu")
	PrintMenu(bot, message, userHandler)
}

func PrintMenu(bot *tgbotapi.BotAPI, message *tgbotapi.Message, userHandler *userDelivery.UserHandler) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	msg.Text = "Вы находитесь в меню. Тут будет несколько кнопок"
	bot.Send(msg)
}
