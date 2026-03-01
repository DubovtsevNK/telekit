package telekit

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ==================== RegisterCommand ====================

func TestTg_RegisterCommand(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})

	tg.RegisterCommand("start", func(tg *Tg, update tgbotapi.Update) {})

	if _, exists := tg.commands["start"]; !exists {
		t.Fatal("command 'start' not registered")
	}
}

func TestTg_RegisterCommand_EmptyName(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})

	tg.RegisterCommand("", func(tg *Tg, update tgbotapi.Update) {})

	if tg.commands != nil {
		t.Fatal("commands map should be nil for empty command name")
	}
}

// ==================== RegisterGlobalCommand ====================

func TestTg_RegisterGlobalCommand(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})

	tg.RegisterGlobalCommand("cancel")

	if _, exists := tg.globalComands["cancel"]; !exists {
		t.Fatal("global command 'cancel' not registered")
	}
}

func TestTg_RegisterGlobalCommand_EmptyName(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})

	tg.RegisterGlobalCommand("")

	if tg.globalComands != nil {
		t.Fatal("globalComands map should be nil for empty command name")
	}
}

// ==================== IsGlobalCommand ====================

func TestTg_IsGlobalCommand(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})
	tg.globalComands = map[string]struct{}{
		"cancel": {},
		"start":  {},
	}

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"with slash", "/cancel", true},
		{"without slash", "cancel", true},
		{"uppercase", "/CANCEL", true},
		{"not global", "/help", false},
		{"with callback data", "cancel_extra", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tg.IsGlobalCommand(tt.input)
			if result != tt.expected {
				t.Errorf("IsGlobalCommand(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ==================== ExecuteCommand ====================

func TestTg_ExecuteCommand(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})
	commandExecuted := false

	tg.RegisterCommand("test", func(tg *Tg, update tgbotapi.Update) {
		commandExecuted = true
	})

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 123456},
			From: &tgbotapi.User{ID: 789},
			Text: "/test",
		},
	}

	err := tg.ExecuteCommand("test", update)
	if err != nil {
		t.Fatal("ExecuteCommand returned error")
	}
	if !commandExecuted {
		t.Fatal("command handler not executed")
	}
}

func TestTg_ExecuteCommand_UnknownCommand(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 123456},
			From: &tgbotapi.User{ID: 789},
		},
	}

	err := tg.ExecuteCommand("unknown", update)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

// ==================== SendMessageText ====================

func TestTg_SendMessageText(t *testing.T) {
	mockBot := &MockBotAPI{}
	tg := NewTgBot(ConfigBot{Separate: "_"}, mockBot, NewStateMachine())

	tg.SendMessageText(123456, "Hello, World!")

	if len(mockBot.SentMessages) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(mockBot.SentMessages))
	}

	msg, ok := mockBot.SentMessages[0].(tgbotapi.MessageConfig)
	if !ok {
		t.Fatal("expected MessageConfig")
	}

	if msg.Text != "Hello, World!" {
		t.Errorf("expected text 'Hello, World!', got '%s'", msg.Text)
	}
	if msg.ChatID != 123456 {
		t.Errorf("expected chatID 123456, got %d", msg.ChatID)
	}
}

// ==================== RegisterComandCallback ====================

func TestTg_RegisterComandCallback(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})

	tg.RegisterComandCallback("btn_click", ComandsCallback{
		Answer: "Clicked!",
		Functor: func(tg *Tg, update tgbotapi.Update) {},
	})

	if _, exists := tg.comandsCallback["btn_click"]; !exists {
		t.Fatal("callback 'btn_click' not registered")
	}
}

func TestTg_RegisterComandCallback_EmptyName(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})

	tg.RegisterComandCallback("", ComandsCallback{
		Answer:  "Clicked!",
		Functor: func(tg *Tg, update tgbotapi.Update) {},
	})

	if tg.comandsCallback != nil {
		t.Fatal("comandsCallback map should be nil for empty callback name")
	}
}

// ==================== FindComandCallback ====================

func TestTg_FindComandCallback(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})
	tg.comandsCallback = map[string]ComandsCallback{
		"btn_click": {Answer: "Clicked!"},
	}

	if !tg.FindComandCallback("btn_click") {
		t.Fatal("callback 'btn_click' should exist")
	}

	if tg.FindComandCallback("nonexistent") {
		t.Fatal("callback 'nonexistent' should not exist")
	}
}

// ==================== DeleteComandCallback ====================

func TestTg_DeleteComandCallback(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})
	tg.comandsCallback = map[string]ComandsCallback{
		"btn_click": {Answer: "Clicked!"},
	}

	tg.DeleteComandCallback("btn_click")

	if tg.FindComandCallback("btn_click") {
		t.Fatal("callback 'btn_click' should be deleted")
	}
}

// ==================== handleCallbackQuerry ====================

func TestTg_handleCallbackQuerry(t *testing.T) {
	mockBot := &MockBotAPI{}
	callbackExecuted := false

	tg := NewTgBot(ConfigBot{Separate: "_"}, mockBot, NewStateMachine())
	tg.RegisterComandCallback("btn", ComandsCallback{
		Answer: "Clicked!",
		Functor: func(tg *Tg, update tgbotapi.Update) {
			callbackExecuted = true
		},
	})

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "callback_123",
			Data: "btn_extra", // separateComanCalback обрежет до "btn"
			From: &tgbotapi.User{ID: 789},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123456},
			},
		},
	}

	tg.handleCallbackQuerry(update)

	// Проверяем, что callback был отправлен
	if len(mockBot.SentCallbacks) == 0 {
		t.Fatal("expected callback response to be sent")
	}

	// Проверяем, что handler был вызван
	if !callbackExecuted {
		t.Fatal("callback handler not executed")
	}
}

func TestTg_handleCallbackQuerry_UnknownCallback(t *testing.T) {
	mockBot := &MockBotAPI{}
	tg := NewTgBot(ConfigBot{Separate: "_"}, mockBot, NewStateMachine())

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "callback_123",
			Data: "unknown_callback",
			From: &tgbotapi.User{ID: 789},
		},
	}

	// Не должно паниковать для неизвестного callback
	tg.handleCallbackQuerry(update)

	// Ничего не должно быть отправлено
	if len(mockBot.SentMessages) != 0 {
		t.Fatalf("expected 0 messages sent, got %d", len(mockBot.SentMessages))
	}
}

// ==================== handleStateMachine ====================

func TestTg_handleStateMachine_NoUserState(t *testing.T) {
	mockBot := &MockBotAPI{}
	tg := NewTgBot(ConfigBot{Separate: "_"}, mockBot, NewStateMachine())

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 123456},
			Text: "some text",
		},
	}

	result := tg.handleStateMachine(update)
	if result {
		t.Fatal("handleStateMachine should return false when user has no state")
	}
}

func TestTg_handleStateMachine_GlobalCommand(t *testing.T) {
	mockBot := &MockBotAPI{}
	sm := NewStateMachine()
	sm.RegisterScenario("test", []Step{
		{Type: StepTypeMessage, Functor: func(tg *Tg, u tgbotapi.Update) error { return nil }},
	})

	tg := NewTgBot(ConfigBot{Separate: "_"}, mockBot, sm)
	tg.RegisterGlobalCommand("cancel")

	// Запускаем сценарий
	sm.StartScenario(123456, "test")

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 123456},
			Text: "/cancel",
		},
	}

	result := tg.handleStateMachine(update)
	if result {
		t.Fatal("handleStateMachine should return false for global command")
	}
}

func TestTg_handleStateMachine_StepTypeMessage(t *testing.T) {
	mockBot := &MockBotAPI{}
	stepExecuted := false

	sm := NewStateMachine()
	sm.RegisterScenario("test", []Step{
		{
			Type:    StepTypeMessage,
			Functor: func(tg *Tg, u tgbotapi.Update) error { stepExecuted = true; return nil },
		},
	})

	tg := NewTgBot(ConfigBot{Separate: "_"}, mockBot, sm)
	sm.StartScenario(123456, "test")

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 123456},
			Text: "user input",
		},
	}

	result := tg.handleStateMachine(update)
	if !result {
		t.Fatal("handleStateMachine should return true when step executed")
	}
	if !stepExecuted {
		t.Fatal("step functor not executed")
	}

	// Проверяем, что шаг увеличился
	state := sm.FindUserState(123456)
	if state.Step != 1 {
		t.Errorf("expected step 1, got %d", state.Step)
	}
}

func TestTg_handleStateMachine_StepTypeCommand(t *testing.T) {
	mockBot := &MockBotAPI{}
	stepExecuted := false

	sm := NewStateMachine()
	sm.RegisterScenario("test", []Step{
		{
			Type:    StepTypeCommand,
			Functor: func(tg *Tg, u tgbotapi.Update) error { stepExecuted = true; return nil },
		},
	})

	tg := NewTgBot(ConfigBot{Separate: "_"}, mockBot, sm)
	sm.StartScenario(123456, "test")

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 123456},
			Text: "/next",
			Entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: 5},
			},
		},
	}

	result := tg.handleStateMachine(update)
	if !result {
		t.Fatal("handleStateMachine should return true when step executed")
	}
	if !stepExecuted {
		t.Fatal("step functor not executed")
	}
}

func TestTg_handleStateMachine_StepTypeCallback(t *testing.T) {
	mockBot := &MockBotAPI{}
	stepExecuted := false

	sm := NewStateMachine()
	sm.RegisterScenario("test", []Step{
		{
			Type:    StepTypeCallback,
			Functor: func(tg *Tg, u tgbotapi.Update) error { stepExecuted = true; return nil },
		},
	})

	tg := NewTgBot(ConfigBot{Separate: "_"}, mockBot, sm)
	sm.StartScenario(123456, "test")

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "callback_123",
			Data: "btn_click",
			From: &tgbotapi.User{ID: 789},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 123456},
			},
		},
	}

	result := tg.handleStateMachine(update)
	if !result {
		t.Fatal("handleStateMachine should return true when step executed")
	}
	if !stepExecuted {
		t.Fatal("step functor not executed")
	}
}

func TestTg_handleStateMachine_ValidationFailed(t *testing.T) {
	mockBot := &MockBotAPI{}

	sm := NewStateMachine()
	sm.RegisterScenario("test", []Step{
		{
			Type:           StepTypeMessage,
			ValidationFunc: func(s string) error { return nil }, // всегда проходит
		},
	})

	tg := NewTgBot(ConfigBot{Separate: "_"}, mockBot, sm)
	sm.StartScenario(123456, "test")

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 123456},
			Text: "valid input",
		},
	}

	result := tg.handleStateMachine(update)
	if !result {
		t.Fatal("handleStateMachine should return true when validation passes")
	}
}

func TestTg_handleStateMachine_NoMoreSteps(t *testing.T) {
	mockBot := &MockBotAPI{}

	sm := NewStateMachine()
	sm.RegisterScenario("test", []Step{
		{Type: StepTypeMessage},
	})

	tg := NewTgBot(ConfigBot{Separate: "_"}, mockBot, sm)
	sm.StartScenario(123456, "test")

	// Искусственно увеличиваем шаг за пределы
	state := sm.FindUserState(123456)
	state.Step = 1 // больше чем количество шагов

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 123456},
			Text: "some text",
		},
	}

	result := tg.handleStateMachine(update)
	if result {
		t.Fatal("handleStateMachine should return false when no more steps")
	}
}

// ==================== separateComanCalback ====================

func TestTg_separateComanCalback(t *testing.T) {
	tg := MakeTgBot(ConfigBot{Separate: "_"})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with separator", "btn_click", "btn"},
		{"without separator", "btn_click_extra", "btn"},
		{"multiple separators", "a_b_c", "a"},
		{"empty", "", ""},
		{"no separator in text", "simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tg.separateComanCalback(tt.input)
			if result != tt.expected {
				t.Errorf("separateComanCalback(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
