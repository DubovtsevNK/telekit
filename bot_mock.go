package telekit

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// MockBotAPI — мок для тестирования BotAPI
type MockBotAPI struct {
	SentMessages      []tgbotapi.Chattable // запоминает все отправленные сообщения
	SentCallbacks     []tgbotapi.CallbackConfig
	SendFunc          func(tgbotapi.Chattable) (tgbotapi.Message, error)
}

// Send имитирует отправку сообщения
func (m *MockBotAPI) Send(chattable tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.SentMessages = append(m.SentMessages, chattable)

	// Сохраняем callback отдельно для удобной проверки
	if callback, ok := chattable.(tgbotapi.CallbackConfig); ok {
		m.SentCallbacks = append(m.SentCallbacks, callback)
	}

	if m.SendFunc != nil {
		return m.SendFunc(chattable)
	}
	return tgbotapi.Message{}, nil
}

// GetUpdatesChan возвращает закрытый канал для тестов
func (m *MockBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	ch := make(chan tgbotapi.Update)
	close(ch)
	return ch
}

// Reset очищает историю сообщений
func (m *MockBotAPI) Reset() {
	m.SentMessages = nil
	m.SentCallbacks = nil
}
