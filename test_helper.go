package telekit

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// TestBot — обёртка для удобного тестирования бота
type TestBot struct {
	TG      *Tg
	MockBot *MockBotAPI
	StateM  *StateMachine
}

// NewTestBot создаёт новый тестовый бот
func NewTestBot() *TestBot {
	return NewTestBotWithConfig(ConfigBot{
		Token:    "test_token",
		Debug:    false,
		Timeout:  60,
		Separate: "_",
	})
}

// NewTestBotWithConfig создаёт тестовый бот с кастомной конфигурацией
func NewTestBotWithConfig(config ConfigBot) *TestBot {
	mockBot := &MockBotAPI{}
	sm := NewStateMachine()
	tg := NewTgBot(config, mockBot, sm)
	return &TestBot{
		TG:      tg,
		MockBot: mockBot,
		StateM:  sm,
	}
}

// Reset очищает состояние тестового бота
func (tb *TestBot) Reset() {
	tb.MockBot.Reset()
	tb.StateM = NewStateMachine()
	tb.TG.StateM = tb.StateM
}

// ==================== Сообщения ====================

// SendMessage эмулирует текстовое сообщение боту
func (tb *TestBot) SendMessage(chatID int64, text string) {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: chatID},
			Text: text,
		},
	}
	tb.TG.handleUpdate(update)
}

// SendMessageWithEntities эмулирует сообщение с entities (например, команды)
func (tb *TestBot) SendMessageWithEntities(chatID int64, text string, entities []tgbotapi.MessageEntity) {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat:     &tgbotapi.Chat{ID: chatID},
			From:     &tgbotapi.User{ID: chatID},
			Text:     text,
			Entities: entities,
		},
	}
	tb.TG.handleUpdate(update)
}

// SendCommand эмулирует команду боту (например, /start)
func (tb *TestBot) SendCommand(chatID int64, command string) {
	text := "/" + command
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: chatID},
			Text: text,
			Entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: len(text)},
			},
		},
	}
	tb.TG.handleUpdate(update)
}

// ==================== Callback ====================

// SendCallback эмулирует нажатие на inline кнопку
func (tb *TestBot) SendCallback(chatID int64, data string) {
	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:    "callback_test",
			Data:  data,
			From:  &tgbotapi.User{ID: chatID},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
			},
		},
	}
	tb.TG.handleUpdate(update)
}

// SendCallbackWithMessage эмулирует callback с полным сообщением
func (tb *TestBot) SendCallbackWithMessage(chatID int64, data string, message *tgbotapi.Message) {
	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:      "callback_test",
			Data:    data,
			From:    &tgbotapi.User{ID: chatID},
			Message: message,
		},
	}
	tb.TG.handleUpdate(update)
}

// ==================== State Machine ====================

// StartScenario запускает сценарий для пользователя
func (tb *TestBot) StartScenario(chatID int64, command string) error {
	return tb.StateM.StartScenario(chatID, command)
}

// GetUserState получает состояние пользователя
func (tb *TestBot) GetUserState(chatID int64) *UserState {
	return tb.StateM.FindUserState(chatID)
}

// GetUserData получает данные пользователя из состояния
func (tb *TestBot) GetUserData(chatID int64) map[string]string {
	state := tb.StateM.FindUserState(chatID)
	if state == nil {
		return nil
	}
	return state.Data
}

// HasActiveState проверяет, есть ли у пользователя активное состояние
func (tb *TestBot) HasActiveState(chatID int64) bool {
	return tb.StateM.FindUserState(chatID) != nil
}

// ==================== Проверки ====================

// GetLastSentMessage возвращает последнее отправленное сообщение
func (tb *TestBot) GetLastSentMessage() *tgbotapi.MessageConfig {
	if len(tb.MockBot.SentMessages) == 0 {
		return nil
	}
	last := tb.MockBot.SentMessages[len(tb.MockBot.SentMessages)-1]
	if msg, ok := last.(tgbotapi.MessageConfig); ok {
		return &msg
	}
	return nil
}

// GetLastSentText возвращает текст последнего отправленного сообщения
func (tb *TestBot) GetLastSentText() string {
	msg := tb.GetLastSentMessage()
	if msg == nil {
		return ""
	}
	return msg.Text
}

// GetLastSentChatID возвращает chatID последнего отправленного сообщения
func (tb *TestBot) GetLastSentChatID() int64 {
	msg := tb.GetLastSentMessage()
	if msg == nil {
		return 0
	}
	return msg.ChatID
}

// GetSentMessagesCount возвращает количество отправленных сообщений
func (tb *TestBot) GetSentMessagesCount() int {
	return len(tb.MockBot.SentMessages)
}

// GetLastCallback возвращает последний отправленный callback
func (tb *TestBot) GetLastCallback() *tgbotapi.CallbackConfig {
	if len(tb.MockBot.SentCallbacks) == 0 {
		return nil
	}
	last := tb.MockBot.SentCallbacks[len(tb.MockBot.SentCallbacks)-1]
	return &last
}

// GetLastCallbackText возвращает текст последнего callback
func (tb *TestBot) GetLastCallbackText() string {
	cb := tb.GetLastCallback()
	if cb == nil {
		return ""
	}
	return cb.Text
}

// WasCommandSent проверяет, было ли отправлено сообщение с командой
func (tb *TestBot) WasCommandSent() bool {
	for _, msg := range tb.MockBot.SentMessages {
		if _, ok := msg.(tgbotapi.MessageConfig); ok {
			return true
		}
	}
	return false
}

// WasCallbackSent проверяет, был ли отправлен callback
func (tb *TestBot) WasCallbackSent() bool {
	return len(tb.MockBot.SentCallbacks) > 0
}

// ClearSentMessages очищает историю отправленных сообщений
func (tb *TestBot) ClearSentMessages() {
	tb.MockBot.Reset()
}

// ==================== Сценарии ====================

// RegisterScenario регистрирует сценарий с шагами
func (tb *TestBot) RegisterScenario(command string, steps []Step) {
	tb.StateM.RegisterScenario(command, steps)
}

// RegisterScenarioWithNames регистрирует сценарий с именованными шагами
func (tb *TestBot) RegisterScenarioWithNames(command string, stepFuncs map[string]func(*Tg, tgbotapi.Update) error) {
	steps := make([]Step, 0, len(stepFuncs))
	for _, fn := range stepFuncs {
		steps = append(steps, Step{
			Type:    StepTypeMessage,
			Functor: fn,
		})
	}
	tb.StateM.RegisterScenario(command, steps)
}

// CompleteScenario завершает сценарий для пользователя
func (tb *TestBot) CompleteScenario(chatID int64) {
	state := tb.StateM.FindUserState(chatID)
	if state != nil {
		tb.StateM.CompleteScenario(chatID, state)
	}
}

// ==================== Commands & Callbacks ====================

// RegisterCommand регистрирует команду
func (tb *TestBot) RegisterCommand(name string, handler func(*Tg, tgbotapi.Update)) {
	tb.TG.RegisterCommand(name, handler)
}

// RegisterGlobalCommand регистрирует глобальную команду
func (tb *TestBot) RegisterGlobalCommand(name string) {
	tb.TG.RegisterGlobalCommand(name)
}

// RegisterCallback регистрирует callback
func (tb *TestBot) RegisterCallback(name string, answer string, handler func(*Tg, tgbotapi.Update)) {
	tb.TG.RegisterComandCallback(name, ComandsCallback{
		Answer:  answer,
		Functor: handler,
	})
}

// HasCommand проверяет, зарегистрирована ли команда
func (tb *TestBot) HasCommand(name string) bool {
	_, ok := tb.TG.commands[name]
	return ok
}

// HasCallback проверяет, зарегистрирован ли callback
func (tb *TestBot) HasCallback(name string) bool {
	return tb.TG.FindComandCallback(name)
}

// IsGlobalCommand проверяет, является ли команда глобальной
func (tb *TestBot) IsGlobalCommand(cmd string) bool {
	return tb.TG.IsGlobalCommand(cmd)
}
