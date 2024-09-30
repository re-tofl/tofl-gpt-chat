package main

import (
	"database/sql"
	"log"
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"tgbot/domain"
	tgDelivery "tgbot/internal/telegramAPI/delivery"
)

func main() {
	logger := CreateLogger()
	tgHandler := tgDelivery.NewTelegramHandler(logger)
	config := loadConfig(logger)
	tgHandler.CreateTelegramBot(config.TgKey)
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

func HandleTelegramMessages(telegramHandler *tgDelivery.TelegramHandler, logger *zap.SugaredLogger, postgresDB *sql.DB) {

	bot := TelegramHandler.TGBot.Bot

	updates, err := bot.GetUpdatesChan(telegramHandler.TGBot.Updates)
	if err != nil {
		log.Fatal(err)
	}

	user := domain.User{}
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if user.ChatID == 0 {
			user.ChatID = update.Message.Chat.ID
			accExists := handlers.UserHandler.CheckAccountExists(user.ChatID)
			if !accExists {
				user = *handlers.UserHandler.Register(update.Message)
			} else {
				user.State = handlers.UserHandler.SetUserState(user.ChatID, "start")
			}
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		user.State = handlers.UserHandler.GetUserState(user.ChatID)
		switch user.State {
		case "start":
			handleStart(bot, update.Message, &handlers, logger)
		case "menu":
			handleMenu(bot, update.Message, &handlers)
		case "choosingDateOfPublication":
			handleChoosingDate(bot, update.Message, &handlers)
		case "getPortfolioOtherUsers":
			err = handleGetPortfolioOtherUsers(bot, update.Message, &handlers)
			if err != nil {
				handlePortfolioParseErr(bot, update.Message, &handlers)
			}
		case "waitForNewSessionID":
			handleUpdateSessionID(bot, update.Message, logger, &handlers, mongoDB, postgresDB)
		case "updateSessionID":
			handleUpdateSessionID(bot, update.Message, logger, &handlers, mongoDB, postgresDB)

		default:
			handleUnknown(bot, update.Message, &handlers)
		}

	}
}
