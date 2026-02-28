package main

import (
	"fmt"
	"strings"

	"github.com/DubovtsevNK/telekit"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func main() {
	// Настройка логгера (опционально)
	logger, _ := zap.NewDevelopment()
	telekit.SetLogger(logger)

	// Конфигурация бота
	config := telekit.ConfigBot{
		Token:    "YOUR_BOT_TOKEN_HERE", // Замените на ваш токен
		Debug:    true,
		Timeout:  60,
		Separate: "@",
	}

	bot := telekit.MakeTgBot(config)
	bot.Init()

	// Регистрация команд
	bot.RegisterCommand("start", handleStart)
	bot.RegisterCommand("help", handleHelp)
	bot.RegisterCommand("cancel", handleCancel)

	// Глобальная команда (работает даже внутри state machine)
	bot.RegisterGlobalCommand("cancel")

	// Регистрация callback с динамическими ID
	// btn_book@1, btn_book@2, btn_book@3 — все будут обрабатываться одной функцией
	bot.RegisterComandCallback("btn_book", telekit.ComandsCallback{
		Answer:  "Book selected!",
		Functor: handleBookButton,
	})

	// Регистрация сценария регистрации (2 шага: имя → email)
	registerSteps := []telekit.Step{
		{
			Prompt: "Введите ваше имя:",
			Type:   telekit.StepTypeMessage,
			Functor: func(tg *telekit.Tg, update tgbotapi.Update) error {
				name := update.Message.Text
				state := tg.StateM.FindUserState(update.Message.Chat.ID)
				state.Data["name"] = name
				tg.SendMessageText(update.Message.Chat.ID, fmt.Sprintf("Приятно познакомиться, %s!\nТеперь введите ваш email:", name))
				return nil
			},
		},
		{
			Prompt: "Введите ваш email:",
			Type:   telekit.StepTypeMessage,
			ValidationFunc: func(input string) error {
				if !strings.Contains(input, "@") {
					return fmt.Errorf("❌ Неверный формат email. Попробуйте ещё раз.")
				}
				return nil
			},
			Functor: func(tg *telekit.Tg, update tgbotapi.Update) error {
				email := update.Message.Text
				state := tg.StateM.FindUserState(update.Message.Chat.ID)
				state.Data["email"] = email
				tg.SendMessageText(update.Message.Chat.ID, fmt.Sprintf("✅ Регистрация завершена!\n\nИмя: %s\nEmail: %s", state.Data["name"], email))
				return nil
			},
		},
	}

	bot.StateM.RegisterScenario("register", registerSteps)

	// Команда для запуска сценария
	bot.RegisterCommand("register", func(tg *telekit.Tg, update tgbotapi.Update) {
		chatID := update.Message.Chat.ID
		tg.StateM.StartScenario(chatID, "register")
		tg.SendMessageText(chatID, "📝 Запуск регистрации...\n\nВведите ваше имя:")
	})

	// Команда для показа кнопок с ID
	bot.RegisterCommand("buttons", func(tg *telekit.Tg, update tgbotapi.Update) {
		chatID := update.Message.Chat.ID
		msg := tgbotapi.NewMessage(chatID, "Выберите книгу:")
		
		// Пример динамических кнопок с ID
		// Все кнопки btn_book@1, btn_book@2, btn_book@3 будут обрабатываться одной функцией handleBookButton
		// В handleBookButton ты можешь получить ID после "@" и действовать соответственно
		btn := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Книга 1", "btn_book@1"),
			tgbotapi.NewInlineKeyboardButtonData("Книга 2", "btn_book@2"),
			tgbotapi.NewInlineKeyboardButtonData("Книга 3", "btn_book@3"),
		)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(btn)
		tg.SendMessage(msg)
	})

	fmt.Println("🤖 Бот запущен...")
	bot.StartListenMessage()
}

// Обработчики команд
func handleStart(tg *telekit.Tg, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	tg.SendMessageText(chatID, "👋 Привет! Я Telekit бот.\n\nДоступные команды:\n/help - помощь\n/register - регистрация\n/buttons - показать кнопки\n/cancel - отменить операцию")
}

func handleHelp(tg *telekit.Tg, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	tg.SendMessageText(chatID, "📖 Помощь:\n\n/start - начать работу\n/register - пройти регистрацию\n/buttons - тест кнопок\n/cancel - отменить текущую операцию")
}

func handleCancel(tg *telekit.Tg, update tgbotapi.Update) {
	var chatID int64
	if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
	} else {
		return
	}

	state := tg.StateM.FindUserState(chatID)
	if state != nil {
		tg.StateM.CompleteScenario(chatID, state)
		tg.SendMessageText(chatID, "❌ Операция отменена.")
	}
}

func handleBookButton(tg *telekit.Tg, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	data := update.CallbackQuery.Data
	
	// Получаем ID книги после сепаратора "@"
	// data = "btn_book@1" → bookID = "1"
	// data = "btn_book@42" → bookID = "42"
	bookID := strings.TrimPrefix(data, "btn_book@")
	
	tg.SendMessageText(chatID, fmt.Sprintf("📚 Вы выбрали книгу #%s", bookID))
}
