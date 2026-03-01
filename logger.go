package telekit

import "go.uber.org/zap"

var log *zap.Logger

func init() {
	var err error
	log, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

// SetLogger allows user to provide custom logger
// Example:
//
//	customLogger, _ := zap.NewDevelopment()
//	telekit.SetLogger(customLogger)
//	bot := telekit.MakeTgBot(config)
//	bot.Init()
func SetLogger(logger *zap.Logger) {
	log = logger
}
