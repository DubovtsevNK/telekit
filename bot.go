package telekit

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type ComandsCallback struct {
	Answer  string
	Functor func(tg *Tg, update tgbotapi.Update)
}

type Tg struct {
	Bot             BotAPI
	StateM          *StateMachine
	commands        map[string]func(tg *Tg, update tgbotapi.Update)
	comandsCallback map[string]ComandsCallback
	globalComands   map[string]struct{} // commands that work even when user is in state machine
	Config          ConfigBot
}

type ConfigBot struct {
	Token    string
	Debug    bool
	Timeout  int
	Separate string
}

func MakeTgBot(config ConfigBot) *Tg {
	return &Tg{Config: config}
}

// NewTgBot создаёт Tg с внедрёнными зависимостями (для тестов)
func NewTgBot(config ConfigBot, bot BotAPI, stateMachine *StateMachine) *Tg {
	return &Tg{
		Config: config,
		Bot:    bot,
		StateM: stateMachine,
	}
}

func (tg *Tg) Init() error {
	botToken := tg.Config.Token

	realBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("failed to connect telegram bot: %w", err)
	}
	realBot.Debug = tg.Config.Debug
	tg.Bot = realBot
	tg.StateM = NewStateMachine()
	log.Info("telegram bot initialized successfully",
		zap.String("bot_username", tg.Bot.(*tgbotapi.BotAPI).Self.UserName))
	return nil
}

func (tg *Tg) StartListenMessage() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = tg.Config.Timeout
	updates := tg.Bot.GetUpdatesChan(u)

	for update := range updates {
		go tg.handleUpdate(update)
	}
}

func (tg *Tg) handleUpdate(update tgbotapi.Update) {
	log.Debug("update received",
		zap.Int("update_id", update.UpdateID),
		zap.Bool("has_message", update.Message != nil),
		zap.Bool("has_callback", update.CallbackQuery != nil))

	if tg.handleStateMachine(update) { // state machine events have highest priority; other commands are ignored when user is in state machine
		return
	}

	if update.CallbackQuery != nil {
		tg.handleCallbackQuerry(update)
		return
	}

	if update.Message == nil {
		return
	}

	if update.Message.IsCommand() {
		err := tg.ExecuteCommand(update.Message.Command(), update)
		if err != nil {
			log.Error("error command",
				zap.String("command name", update.Message.Command()))
		}
		return
	}
}

func (tg *Tg) ExecuteCommand(cmd string, update tgbotapi.Update) error {
	if handler, exists := tg.commands[cmd]; exists {
		handler(tg, update)
		return nil
	}
	log.Warn("unknown command",
		zap.String("command", cmd),
		zap.Int64("user_id", update.Message.From.ID))
	return fmt.Errorf("unknown command")
}

func (tg *Tg) SendMessageText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := tg.Bot.Send(msg); err != nil {
		log.Fatal("TGBOT : send message",
			zap.String("error", err.Error()))
	}
	log.Debug("Send Message", zap.Int64("UserID ", chatID), zap.String("Text", text))
}

func (tg *Tg) SendMessage(msg tgbotapi.MessageConfig) {
	if _, err := tg.Bot.Send(msg); err != nil {
		log.Error("failed to send message",
			zap.String("error", err.Error()),
			zap.Int64("chat_id", msg.ChatID))
		log.Fatal(err.Error())
	}
	log.Debug("message sent",
		zap.Int64("chat_id", msg.ChatID))
}

func (tg *Tg) RegisterCommand(name string, handler func(tg *Tg, update tgbotapi.Update)) {
	if name == "" {
		return
	}

	if tg.commands == nil {
		tg.commands = make(map[string]func(tg *Tg, update tgbotapi.Update))
		log.Info("initializing commands map")
	}
	tg.commands[name] = handler
	log.Info("command registered",
		zap.String("command", name))
}

func (tg *Tg) RegisterComandCallback(name string, cmdclb ComandsCallback) {
	if name == "" {
		return
	}

	if tg.comandsCallback == nil {
		tg.comandsCallback = make(map[string]ComandsCallback)
		log.Info("initializing callback map")
	}
	tg.comandsCallback[name] = cmdclb
	log.Info("callback registered",
		zap.String("callback", name),
		zap.String("answer", cmdclb.Answer))
}

func (tg *Tg) RegisterGlobalCommand(command string) {
	if command == "" {
		return
	}

	if tg.globalComands == nil {
		tg.globalComands = make(map[string]struct{})
		log.Info("initializing globalComand map")
	}

	tg.globalComands[command] = struct{}{}
	log.Info("commandGlobal registered",
		zap.String("command", command))
}

func (tg *Tg) IsGlobalCommand(cmd string) bool {
	cmd = strings.TrimPrefix(strings.ToLower(cmd), "/")
	cmd = tg.separateComanCalback(cmd)
	_, ok := tg.globalComands[strings.ToLower(cmd)]
	return ok
}

func (tg *Tg) handleStateMachine(update tgbotapi.Update) bool {

	var chatID int64
	var text string

	if update.Message != nil {
		chatID = update.Message.Chat.ID
		text = update.Message.Text
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		text = update.CallbackQuery.Data
	} else {
		return false // unknown type - not for state machine
	}

	if tg.IsGlobalCommand(text) {
		return false
	}

	userState := tg.StateM.FindUserState(chatID)
	if userState == nil {
		return false
	}
	log.Debug("processing state machine step",
		zap.String("command", userState.CurrentCommand),
		zap.Int("step", userState.Step),
		zap.Int64("user_id", chatID))

	if userState.Step >= tg.StateM.GetCountScenarioStepInComand(userState.CurrentCommand) {
		return false // no more steps
	}

	step := tg.StateM.GetScenarioStep(userState.CurrentCommand, userState.Step)
	if step == nil {
		return false
	}

	switch step.Type {
	case StepTypeMessage:
		if update.Message == nil || update.Message.IsCommand() {
			return false // expecting regular message, but received command or callback
		}
		if step.ValidationFunc != nil {
			if err := step.ValidationFunc(update.Message.Text); err != nil {
				tg.SendMessageText(chatID, err.Error())
				return true
			}
		}

	case StepTypeCommand:
		if update.Message == nil || !update.Message.IsCommand() {
			return false // expecting command, but received something else
		}
		// command validation is optional (e.g., only /cancel allowed)

	case StepTypeCallback:
		if update.CallbackQuery == nil {
			return false // expecting callback, but received message
		}
		// callback_data validation is optional

	default:
		log.Warn("unknown step type", zap.Int("type", int(step.Type)))
		return false
	}

	if step.Functor != nil {
		err := step.Functor(tg, update)
		if err != nil {
			log.Error("step functor failed", zap.Error(err))
			tg.SendMessageText(chatID, "An error occurred. Please try again.")
			return true
		}
	}

	// Move to next step
	userState.Step++
	return true
}

func (tg *Tg) separateComanCalback(comand string) string {

	pos := strings.Index(comand, tg.Config.Separate)
	if pos >= 0 {
		return comand[0:pos]
	}
	return comand
}

func (tg *Tg) handleCallbackQuerry(update tgbotapi.Update) {
	comand := tg.separateComanCalback(update.CallbackQuery.Data)
	cmd, exist := tg.comandsCallback[comand]
	if !exist {
		return
	}
	log.Debug("callback query received",
		zap.String("callback_data", update.CallbackQuery.Data),
		zap.Int64("user_id", update.CallbackQuery.From.ID))
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, cmd.Answer) // respond to callback first
	if _, err := tg.Bot.Send(callback); err != nil {
		log.Error("failed to send callback answer",
			zap.String("error", err.Error()),
			zap.Int64("user_id", update.CallbackQuery.From.ID))
	}
	cmd.Functor(tg, update)
}

func (tg *Tg) FindComandCallback(comand string) bool {
	_, err := tg.comandsCallback[comand]
	return err
}

func (tg *Tg) DeleteComandCallback(key string) {
	delete(tg.comandsCallback, key)
}
