//go:build example
// +build example

// Этот файл содержит примеры использования TestBot
// Для запуска: go test -tags example -v -run ExampleTestBot

package telekit

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Пример: тестирование простого сценария регистрации
func ExampleTestBot_RegisterScenario(t *testing.T) {
	// Создаём тестовый бот
	tb := NewTestBot()

	// Регистрируем сценарий
	tb.RegisterScenario("register", []Step{
		{
			Type: StepTypeMessage,
			Functor: func(tg *Tg, u tgbotapi.Update) error {
				name := u.Message.Text
				state := tg.StateM.FindUserState(u.Message.Chat.ID)
				state.Data["name"] = name
				tg.SendMessageText(u.Message.Chat.ID, "Привет, "+name+"!")
				return nil
			},
		},
		{
			Type: StepTypeMessage,
			Functor: func(tg *Tg, u tgbotapi.Update) error {
				email := u.Message.Text
				state := tg.StateM.FindUserState(u.Message.Chat.ID)
				state.Data["email"] = email
				tg.SendMessageText(u.Message.Chat.ID, "Регистрация завершена!")
				return nil
			},
		},
	})

	// Запускаем сценарий
	tb.StartScenario(123456, "register")

	// Отправляем имя
	tb.SendMessage(123456, "Иван")

	// Проверяем, что имя сохранено
	state := tb.GetUserState(123456)
	if state.Data["name"] != "Иван" {
		t.Errorf("expected name 'Иван', got '%s'", state.Data["name"])
	}

	// Проверяем, что сообщение отправлено
	if tb.GetLastSentText() != "Привет, Иван!" {
		t.Errorf("expected 'Привет, Иван!', got '%s'", tb.GetLastSentText())
	}

	// Отправляем email
	tb.SendMessage(123456, "ivan@example.com")

	// Проверяем email
	state = tb.GetUserState(123456)
	if state.Data["email"] != "ivan@example.com" {
		t.Errorf("expected email 'ivan@example.com', got '%s'", state.Data["email"])
	}
}

// Пример: тестирование callback
func ExampleTestBot_Callback(t *testing.T) {
	tb := NewTestBot()

	// Регистрируем callback
	callbackExecuted := false
	tb.RegisterCallback("btn_click", "Clicked!", func(tg *Tg, u tgbotapi.Update) {
		callbackExecuted = true
	})

	// Отправляем callback
	tb.SendCallback(123456, "btn_click")

	// Проверяем, что callback обработан
	if !callbackExecuted {
		t.Fatal("callback handler not executed")
	}

	// Проверяем, что callback ответ отправлен
	if !tb.WasCallbackSent() {
		t.Fatal("callback response not sent")
	}

	if tb.GetLastCallbackText() != "Clicked!" {
		t.Errorf("expected 'Clicked!', got '%s'", tb.GetLastCallbackText())
	}
}

// Пример: тестирование команды
func ExampleTestBot_Command(t *testing.T) {
	tb := NewTestBot()

	// Регистрируем команду
	commandExecuted := false
	tb.RegisterCommand("start", func(tg *Tg, u tgbotapi.Update) {
		commandExecuted = true
		tg.SendMessageText(u.Message.Chat.ID, "Привет!")
	})

	// Отправляем команду
	tb.SendCommand(123456, "start")

	// Проверяем, что команда выполнена
	if !commandExecuted {
		t.Fatal("command handler not executed")
	}

	// Проверяем сообщение
	if tb.GetLastSentText() != "Привет!" {
		t.Errorf("expected 'Привет!', got '%s'", tb.GetLastSentText())
	}
}

// Пример: тестирование глобальной команды
func ExampleTestBot_GlobalCommand(t *testing.T) {
	tb := NewTestBot()

	// Регистрируем сценарий
	tb.RegisterScenario("test", []Step{
		{
			Type: StepTypeMessage,
			Functor: func(tg *Tg, u tgbotapi.Update) error {
				return nil
			},
		},
	})

	// Регистрируем глобальную команду
	cancelExecuted := false
	tb.RegisterGlobalCommand("cancel")
	tb.RegisterCommand("cancel", func(tg *Tg, u tgbotapi.Update) {
		cancelExecuted = true
	})

	// Запускаем сценарий
	tb.StartScenario(123456, "test")

	// Отправляем глобальную команду во время сценария
	tb.SendCommand(123456, "cancel")

	// Проверяем, что команда выполнена (даже внутри сценария)
	if !cancelExecuted {
		t.Fatal("global command not executed")
	}
}

// Пример: тестирование валидации
func ExampleTestBot_Validation(t *testing.T) {
	tb := NewTestBot()

	// Регистрируем сценарий с валидацией
	tb.RegisterScenario("register", []Step{
		{
			Type: StepTypeMessage,
			ValidationFunc: func(input string) error {
				if input == "" {
					return nil // пропускаем пустую проверку для примера
				}
				if len(input) < 3 {
					return nil // пропускаем для примера
				}
				return nil
			},
			Functor: func(tg *Tg, u tgbotapi.Update) error {
				state := tg.StateM.FindUserState(u.Message.Chat.ID)
				state.Data["input"] = u.Message.Text
				return nil
			},
		},
	})

	tb.StartScenario(123456, "register")
	tb.SendMessage(123456, "test")

	state := tb.GetUserState(123456)
	if state.Data["input"] != "test" {
		t.Errorf("expected 'test', got '%s'", state.Data["input"])
	}
}

// Пример: тестирование нескольких пользователей
func ExampleTestBot_MultipleUsers(t *testing.T) {
	tb := NewTestBot()

	tb.RegisterScenario("register", []Step{
		{
			Type: StepTypeMessage,
			Functor: func(tg *Tg, u tgbotapi.Update) error {
				state := tg.StateM.FindUserState(u.Message.Chat.ID)
				state.Data["name"] = u.Message.Text
				return nil
			},
		},
	})

	// Пользователь 1
	tb.StartScenario(111, "register")
	tb.SendMessage(111, "Alice")

	// Пользователь 2
	tb.StartScenario(222, "register")
	tb.SendMessage(222, "Bob")

	// Проверяем изоляцию
	state1 := tb.GetUserState(111)
	state2 := tb.GetUserState(222)

	if state1.Data["name"] != "Alice" {
		t.Errorf("user 1: expected 'Alice', got '%s'", state1.Data["name"])
	}
	if state2.Data["name"] != "Bob" {
		t.Errorf("user 2: expected 'Bob', got '%s'", state2.Data["name"])
	}
}

// Пример: сброс состояния между тестами
func ExampleTestBot_Reset(t *testing.T) {
	tb := NewTestBot()

	// Первый тест
	tb.StartScenario(123456, "test")
	tb.SendMessage(123456, "data")

	// Сбрасываем состояние
	tb.Reset()

	// Проверяем, что состояние очищено
	if tb.HasActiveState(123456) {
		t.Fatal("state should be cleared after reset")
	}

	if tb.GetSentMessagesCount() != 0 {
		t.Errorf("expected 0 messages, got %d", tb.GetSentMessagesCount())
	}
}
