package telekit

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// ==================== Bot Types ====================

// Config represents Telegram bot configuration
type Config struct {
	Token    string // Telegram bot token
	Debug    bool   // Enable debug mode
	Timeout  int    // Update timeout in seconds
	Separate string // Separator for callback data (default "_")
}

// CommandCallback represents callback query handler
type CommandCallback struct {
	Answer  string                               // Callback response text
	Functor func(tg *Tg, update tgbotapi.Update) // Handler function
}

// ==================== State Machine Types ====================

// StepType defines state machine step type
type StepType int

const (
	StepTypeMessage  StepType = iota // Regular text message
	StepTypeCommand                  // Command (e.g., /start, /cancel)
	StepTypeCallback                 // Callback query (inline button)
)

// Step describes one scenario step
type Step struct {
	Prompt         string                           // User prompt (optional)
	Functor        func(*Tg, tgbotapi.Update) error // Step execution function
	ValidationFunc func(string) error               // Input validation function (optional)
	Type           StepType                         // Step type
}

// Scenario represents execution scenario
type Scenario struct {
	Steps []Step // Scenario steps
}

// UserState represents user state in state machine
type UserState struct {
	CurrentCommand string            // Current executing command
	Step           int               // Current step in scenario
	Data           map[string]string // Temporary user data
}

// StateMachine manages user states
type StateMachine struct {
	userStates map[int64]*UserState // chatID -> UserState
	scenarios  map[string]Scenario  // command -> scenario
}
