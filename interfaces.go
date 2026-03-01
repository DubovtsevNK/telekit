package telekit

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// BotAPI определяет методы Telegram Bot API, используемые в проекте
type BotAPI interface {
	Send(chattable tgbotapi.Chattable) (tgbotapi.Message, error)
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}
