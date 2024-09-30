package domain

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type Telegram struct {
	ChatID  int64
	Bot     *tgbotapi.BotAPI
	Updates tgbotapi.UpdateConfig
}
