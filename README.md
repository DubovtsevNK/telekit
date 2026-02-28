# Telekit

A Go library for building Telegram bots with state machine support.

Telekit simplifies the development of Telegram bots by providing a built-in state machine for managing multi-step dialogues and complex conversation flows.

## Features

- **State Machine** — Built-in state machine for managing multi-step dialogues
- **Scenarios** — Define complex conversation flows with ease
- **Input Validation** — Validate user input at each step
- **Callback Support** — Handle inline button callbacks with data parsing
- **Command Handling** — Register and handle bot commands
- **Logging** — Structured logging with Zap (customizable)
- **Lightweight** — Minimal dependencies, easy to integrate

## Installation

```bash
go get github.com/DubovtsevNK/telekit
```

## Examples

See the [examples](examples/basic_bot) directory for a complete working bot example.

## Quick Start

```go
package main

import (
    "github.com/DubovtsevNK/telekit"
    "go.uber.org/zap"
)

func main() {
    // Optional: set custom logger
    logger, _ := zap.NewDevelopment()
    telekit.SetLogger(logger)

    // Create bot
    config := telekit.ConfigBot{
        Token:    "YOUR_BOT_TOKEN",
        Debug:    true,
        Timeout:  60,
        Separate: "_",
    }
    bot := telekit.MakeTgBot(config)
    bot.Init()

    // Register commands
    bot.RegisterCommand("start", handleStart)
    bot.RegisterCommand("help", handleHelp)

    // Start listening
    bot.StartListenMessage()
}

func handleStart(bot *telekit.Tg, update tgbotapi.Update) {
    bot.SendMessageText(update.Message.Chat.ID, "Hello! Welcome to the bot.")
}

func handleHelp(bot *telekit.Tg, update tgbotapi.Update) {
    bot.SendMessageText(update.Message.Chat.ID, "Here's how to use me...")
}
```

## State Machine

Telekit includes a built-in state machine for multi-step dialogues.

### Registering a Scenario

```go
// Define steps
steps := []telekit.Step{
    {
        Prompt: "Please enter your name:",
        Type:   telekit.StepTypeMessage,
        Functor: func(tg *telekit.Tg, update tgbotapi.Update) error {
            name := update.Message.Text
            // Store name in user state
            state := tg.StateM.FindUserState(update.Message.Chat.ID)
            state.Data["name"] = name
            return nil
        },
    },
    {
        Prompt: "Enter your email:",
        Type:   telekit.StepTypeMessage,
        ValidationFunc: func(input string) error {
            if !isValidEmail(input) {
                return fmt.Errorf("invalid email format")
            }
            return nil
        },
        Functor: func(tg *telekit.Tg, update tgbotapi.Update) error {
            email := update.Message.Text
            state := tg.StateM.FindUserState(update.Message.Chat.ID)
            state.Data["email"] = email
            return nil
        },
    },
}

// Register scenario
bot.StateM.RegisterScenario("register", steps)

// Command to start the scenario
bot.RegisterCommand("register", func(bot *telekit.Tg, update tgbotapi.Update) {
    bot.StateM.StartScenario(update.Message.Chat.ID, "register")
    bot.SendMessageText(update.Message.Chat.ID, "Please enter your name:")
})
```

### Step Types

| Type | Description |
|------|-------------|
| `StepTypeMessage` | Regular text message |
| `StepTypeCommand` | Bot command (e.g., `/start`, `/cancel`) |
| `StepTypeCallback` | Inline button callback |

### Global Commands

Global commands work even when the user is in a state machine:

```go
bot.RegisterGlobalCommand("cancel")
bot.RegisterCommand("cancel", func(bot *telekit.Tg, update tgbotapi.Update) {
    state := bot.StateM.FindUserState(update.Message.Chat.ID)
    if state != nil {
        bot.StateM.CompleteScenario(update.Message.Chat.ID, state)
        bot.SendMessageText(update.Message.Chat.ID, "Operation cancelled.")
    }
})
```

## Callback Handling

### Basic Callback

```go
bot.RegisterComandCallback("button_click", telekit.ComandsCallback{
    Answer: "Processing...",
    Functor: func(bot *telekit.Tg, update tgbotapi.Update) {
        data := update.CallbackQuery.Data
        bot.SendMessageText(update.CallbackQuery.Message.Chat.ID, "You clicked: "+data)
    },
})
```

### Callback with Separator (Dynamic IDs)

The `Separate` config option allows you to handle dynamic callbacks with IDs:

```go
// Configure separator (e.g., "@")
config := telekit.ConfigBot{
    Token:    "YOUR_BOT_TOKEN",
    Separate: "@",  // btn_book@1, btn_book@42, etc.
}

// Register callback handler (handles all btn_book@* callbacks)
bot.RegisterComandCallback("btn_book", telekit.ComandsCallback{
    Answer: "Book selected!",
    Functor: func(tg *telekit.Tg, update tgbotapi.Update) {
        data := update.CallbackQuery.Data  // "btn_book@42"
        bookID := strings.TrimPrefix(data, "btn_book@")  // "42"
        // Use bookID to work with specific book
        tg.SendMessageText(update.CallbackQuery.Message.Chat.ID, 
            fmt.Sprintf("Selected book #%s", bookID))
    },
})

// Create inline buttons with dynamic IDs
btn := tgbotapi.NewInlineKeyboardRow(
    tgbotapi.NewInlineKeyboardButtonData("Book 1", "btn_book@1"),
    tgbotapi.NewInlineKeyboardButtonData("Book 2", "btn_book@2"),
    tgbotapi.NewInlineKeyboardButtonData("Book 3", "btn_book@3"),
)
```

**How separator works:**

```
btn_book@1    → handler: "btn_book", ID: "1"
btn_book@42   → handler: "btn_book", ID: "42"
btn_item@123  → handler: "btn_item", ID: "123"
```

This is useful for dynamic buttons: books, products, users, orders, etc.

## Configuration

### ConfigBot

| Field | Type | Description |
|-------|------|-------------|
| `Token` | `string` | Telegram bot token |
| `Debug` | `bool` | Enable debug mode |
| `Timeout` | `int` | Update timeout in seconds |
| `Separate` | `string` | Separator for callback data (default: `"_"`) |

## Logging

By default, Telekit uses a production Zap logger. You can provide a custom logger:

```go
// Development logger with debug output
logger, _ := zap.NewDevelopment()
telekit.SetLogger(logger)

// Or use your own configured logger
customLogger := zap.NewExample()
telekit.SetLogger(customLogger)
```

## License

MIT
