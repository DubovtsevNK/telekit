# Basic Bot Example

Простой пример бота с использованием Telekit.

## Запуск

1. Замените `YOUR_BOT_TOKEN_HERE` в файле `main.go` на ваш токен от @BotFather

2. Инициализируйте зависимости:
```bash
go mod tidy
```

3. Запустите бота:
```bash
go run main.go
```

## Доступные команды

| Команда | Описание |
|---------|----------|
| `/start` | Приветственное сообщение |
| `/help` | Показать справку |
| `/register` | Запустить сценарий регистрации (state machine) |
| `/buttons` | Показать inline кнопки с ID (тест сепаратора) |
| `/cancel` | Отменить текущую операцию (глобальная команда) |

## Особенности примера

- **State Machine** — сценарий регистрации с 2 шагами (имя → email)
- **Валидация** — проверка формата email
- **Callback с сепаратором** — обработка inline кнопок с динамическими ID (`btn_hello@1`, `btn_hello@2`, ...)
- **Глобальные команды** — `/cancel` работает внутри state machine
- **Кастомный логгер** — используется `zap.NewDevelopment()`

## Как работает сепаратор

Сепаратор (`@` в этом примере) разделяет имя callback и его параметры:

```
btn_book@1    → имя: "btn_book", ID: "1"
btn_book@42   → имя: "btn_book", ID: "42" (например, ID книги)
btn_item@123  → имя: "btn_item", ID: "123"
```

В обработчике `handleBookButton` ты получаешь полный `data` и можешь извлечь ID:

```go
func handleBookButton(tg *telekit.Tg, update tgbotapi.Update) {
    data := update.CallbackQuery.Data  // "btn_book@42"
    bookID := strings.TrimPrefix(data, "btn_book@")  // "42"
    // Используй ID для работы с книгой, товаром и т.д.
}
```

Это удобно для динамических кнопок: список книг, товаров, пользователей и т.п.
