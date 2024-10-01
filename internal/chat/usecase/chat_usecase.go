package usecase

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"tgbot/domain"
)

type TgStore interface {
}

func CreateNewTgBot(tgKey string) (*domain.Telegram, error) {
	var telegramBotStruct domain.Telegram
	fmt.Println(tgKey)
	bot, err := tgbotapi.NewBotAPI(tgKey)
	if err != nil {
		return nil, err
	}
	telegramBotStruct.Bot = bot

	return &telegramBotStruct, nil
}
