# TestBot — Утилита для тестирования ботов на telekit

`TestBot` — это обёртка над `Tg` для удобного тестирования ваших сценариев, команд и callback'ов.

---

## 🚀 Быстрый старт

```go
package main

import (
    "testing"
    "github.com/DubovtsevNK/telekit"
)

func TestMyScenario(t *testing.T) {
    // 1. Создаём тестовый бот
    tb := telekit.NewTestBot()
    
    // 2. Регистрируем сценарий
    tb.RegisterScenario("register", []telekit.Step{
        {
            Type: telekit.StepTypeMessage,
            Functor: func(tg *telekit.Tg, u tgbotapi.Update) error {
                name := u.Message.Text
                state := tg.StateM.FindUserState(u.Message.Chat.ID)
                state.Data["name"] = name
                return nil
            },
        },
    })
    
    // 3. Запускаем сценарий
    tb.StartScenario(123456, "register")
    
    // 4. Отправляем данные
    tb.SendMessage(123456, "Иван")
    
    // 5. Проверяем результат
    state := tb.GetUserState(123456)
    if state.Data["name"] != "Иван" {
        t.Errorf("expected 'Иван', got '%s'", state.Data["name"])
    }
}
```

---

## 📋 API

### Создание бота

```go
// Базовое создание
tb := telekit.NewTestBot()

// С кастомной конфигурацией
tb := telekit.NewTestBotWithConfig(telekit.ConfigBot{
    Separate: "@",
    Debug:    false,
})
```

### Отправка сообщений

```go
// Текстовое сообщение
tb.SendMessage(chatID, "привет")

// Команда (например, /start)
tb.SendCommand(chatID, "start")

// Callback (нажатие на кнопку)
tb.SendCallback(chatID, "btn_click")
```

### Проверка результатов

```go
// Последнее отправленное сообщение
text := tb.GetLastSentText()
chatID := tb.GetLastSentChatID()

// Количество сообщений
count := tb.GetSentMessagesCount()

// Callback
cbText := tb.GetLastCallbackText()
wasCallback := tb.WasCallbackSent()

// Состояние пользователя
state := tb.GetUserState(chatID)
data := tb.GetUserData(chatID)
hasState := tb.HasActiveState(chatID)
```

### Управление состоянием

```go
// Сброс состояния (для чистоты между тестами)
tb.Reset()

// Завершение сценария
tb.CompleteScenario(chatID)

// Очистка сообщений
tb.ClearSentMessages()
```

---

## 📖 Примеры

### 1. Тестирование команды

```go
func TestStartCommand(t *testing.T) {
    tb := telekit.NewTestBot()
    
    tb.RegisterCommand("start", func(tg *telekit.Tg, u tgbotapi.Update) {
        tg.SendMessageText(u.Message.Chat.ID, "Привет!")
    })
    
    tb.SendCommand(123456, "start")
    
    if tb.GetLastSentText() != "Привет!" {
        t.Errorf("unexpected message: %s", tb.GetLastSentText())
    }
}
```

### 2. Тестирование сценария с несколькими шагами

```go
func TestRegistrationScenario(t *testing.T) {
    tb := telekit.NewTestBot()
    
    tb.RegisterScenario("register", []telekit.Step{
        {
            Type: telekit.StepTypeMessage,
            Functor: func(tg *telekit.Tg, u tgbotapi.Update) error {
                state := tg.StateM.FindUserState(u.Message.Chat.ID)
                state.Data["name"] = u.Message.Text
                tg.SendMessageText(u.Message.Chat.ID, "Введите email:")
                return nil
            },
        },
        {
            Type: telekit.StepTypeMessage,
            Functor: func(tg *telekit.Tg, u tgbotapi.Update) error {
                state := tg.StateM.FindUserState(u.Message.Chat.ID)
                state.Data["email"] = u.Message.Text
                tg.SendMessageText(u.Message.Chat.ID, "Готово!")
                return nil
            },
        },
    })
    
    tb.StartScenario(123456, "register")
    
    // Шаг 1: имя
    tb.SendMessage(123456, "Иван")
    if tb.GetLastSentText() != "Введите email:" {
        t.Error("expected email prompt")
    }
    
    // Шаг 2: email
    tb.SendMessage(123456, "ivan@example.com")
    if tb.GetLastSentText() != "Готово!" {
        t.Error("expected completion message")
    }
    
    // Проверяем данные
    data := tb.GetUserData(123456)
    if data["name"] != "Иван" {
        t.Error("wrong name")
    }
    if data["email"] != "ivan@example.com" {
        t.Error("wrong email")
    }
}
```

### 3. Тестирование callback

```go
func TestBookCallback(t *testing.T) {
    tb := telekit.NewTestBot()
    
    tb.RegisterCallback("book_select", "Выбрано!", func(tg *telekit.Tg, u tgbotapi.Update) {
        tg.SendMessageText(u.CallbackQuery.Message.Chat.ID, "Книга выбрана")
    })
    
    tb.SendCallback(123456, "book_select_1")
    
    if !tb.WasCallbackSent() {
        t.Fatal("callback not sent")
    }
    if tb.GetLastCallbackText() != "Выбрано!" {
        t.Error("wrong callback text")
    }
}
```

### 4. Тестирование глобальной команды

```go
func TestGlobalCancelCommand(t *testing.T) {
    tb := telekit.NewTestBot()
    
    // Сценарий
    tb.RegisterScenario("order", []telekit.Step{
        {Type: telekit.StepTypeMessage},
    })
    
    // Глобальная команда
    tb.RegisterGlobalCommand("cancel")
    tb.RegisterCommand("cancel", func(tg *telekit.Tg, u tgbotapi.Update) {
        tg.StateM.CompleteScenario(u.Message.Chat.ID, 
            tg.StateM.FindUserState(u.Message.Chat.ID))
    })
    
    // Запускаем сценарий
    tb.StartScenario(123456, "order")
    
    // Отменяем во время сценария
    tb.SendCommand(123456, "cancel")
    
    // Проверяем, что сценарий завершён
    if tb.HasActiveState(123456) {
        t.Error("state should be cleared")
    }
}
```

### 5. Тестирование нескольких пользователей

```go
func TestMultipleUsers(t *testing.T) {
    tb := telekit.NewTestBot()
    
    tb.RegisterScenario("register", []telekit.Step{
        {
            Type: telekit.StepTypeMessage,
            Functor: func(tg *telekit.Tg, u tgbotapi.Update) error {
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
    if tb.GetUserData(111)["name"] != "Alice" {
        t.Error("user 1 data wrong")
    }
    if tb.GetUserData(222)["name"] != "Bob" {
        t.Error("user 2 data wrong")
    }
}
```

---

## 💡 Советы

1. **Используйте `tb.Reset()`** между тестами для чистоты состояния
2. **Тестируйте через `handleUpdate`** — это эмулирует реальную работу бота
3. **Выносите логику** из functor'ов в отдельные функции для unit-тестов
4. **Проверяйте и состояние, и сообщения** — оба аспекта важны

---

## 📁 Файлы

| Файл | Описание |
|------|----------|
| `test_helper.go` | Основной класс `TestBot` |
| `test_helper_example_test.go` | Примеры использования (build tag: `example`) |

---

## 🔗 См. также

- [bot_test.go](bot_test.go) — тесты для `Tg`
- [stateMachine_test.go](stateMachine_test.go) — тесты для `StateMachine`
