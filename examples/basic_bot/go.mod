module github.com/DubovtsevNK/telekit/examples/basic_bot

go 1.25.4

require (
	github.com/DubovtsevNK/telekit v0.0.0
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	go.uber.org/zap v1.27.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/DubovtsevNK/telekit => ../..
