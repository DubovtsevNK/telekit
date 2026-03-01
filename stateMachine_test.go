package telekit

import (
	"reflect"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestStateMachine_FindUserState_Nil(t *testing.T) {

	sm := NewStateMachine()
	chatId := int64(1000)

	state := sm.FindUserState(chatId)

	if state != nil {
		t.Fatal("state not nil")
	}

}

func TestStateMachine_FindUserState(t *testing.T) {

	sm := NewStateMachine()
	chatId := int64(1000)
	sm.RegisterScenario("register", []Step{})
	sm.StartScenario(chatId, "register")
	state := sm.FindUserState(chatId)

	if state == nil {
		t.Fatal("state nill")
	}

	if state.CurrentCommand != "register" {
		t.Errorf("expected command 'register', got '%s'", state.CurrentCommand)
	}

	if state.Step != 0 {
		t.Fatalf("expected step 0, got %d", state.Step)
	}

}

func TestStateMachine_CountCommand(t *testing.T) {

	sm := NewStateMachine()
	countCommand := sm.GetCountScenarioStepInComand("test")

	if countCommand != 0 {
		t.Fatal("expected 0 steps for unregistered command")
	}
	steps := []Step{
		{
			Prompt: "step1",
			Functor: func(t *Tg, u tgbotapi.Update) error {
				return nil
			},
		},
		{
			Prompt: "step2",
			Functor: func(t *Tg, u tgbotapi.Update) error {
				return nil
			},
		},
	}

	sm.RegisterScenario("test", steps)

	countCommand = sm.GetCountScenarioStepInComand("test")

	if countCommand != 2 {
		t.Fatal("expected 2 steps for registered command")
	}
}

func TestStateMachine_CompleteScenario(t *testing.T) {
	sm := NewStateMachine()
	chatId := int64(1000)
	sm.RegisterScenario("register", []Step{})
	sm.StartScenario(chatId, "register")
	state := sm.FindUserState(chatId)

	if state == nil {
		t.Fatal("state not nil")
	}

	sm.CompleteScenario(chatId, state)

	state = sm.FindUserState(chatId)

	if state != nil {
		t.Fatal("user state should be nil after complete")
	}

}

func TestStateMachine_RegisterScenario(t *testing.T) {
	sm := NewStateMachine()

	stepfunc1 := func(t *Tg, u tgbotapi.Update) error {
		return nil
	}

	stepfunc2 := func(t *Tg, u tgbotapi.Update) error {
		return nil
	}

	step1 := Step{
		Prompt:  "step1",
		Functor: stepfunc1,
	}

	step2 := Step{
		Prompt:  "step2",
		Functor: stepfunc2,
	}

	steps := []Step{step1, step2}

	sm.RegisterScenario("test", steps)

	scenario, exists := sm.scenarios["test"]
	if !exists {
		t.Fatal("scenario 'test' not found")
	}

	if len(scenario.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(scenario.Steps))
	}

	equalStep := func(step *Step, equalstep *Step) (bool, string) {
		if step.Prompt != equalstep.Prompt {
			return false, "not equal prompt"
		}

		if step.Type != equalstep.Type {
			return false, "not equal type"
		}

		if reflect.ValueOf(step.Functor).Pointer() != reflect.ValueOf(equalstep.Functor).Pointer() {
			return false, "not equal functor"
		}

		return true, ""
	}

	if ok, msg := equalStep(&scenario.Steps[0], &step1); !ok {
		t.Errorf("step1 not equal %s", msg)
	}

	if ok, msg := equalStep(&scenario.Steps[1], &step2); !ok {
		t.Errorf("step2 not equal %s", msg)
	}
}

func TestStateMachine_StartScenario_NonExistent(t *testing.T) {
	sm := NewStateMachine()
	chatId := int64(1000)

	err := sm.StartScenario(chatId, "register")

	if err == nil {
		t.Fatal("expected error for non-existent scenario")
	}

	state := sm.FindUserState(chatId)
	if state != nil {
		t.Fatal("state should be nil when scenario registration fails")
	}
}

func TestStateMachine_CompleteScenario_NonExistentUser(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CompleteScenario panicked: %v", r)
		}
	}()

	sm := NewStateMachine()

	// Завершаем состояние для пользователя, которого нет
	sm.CompleteScenario(1111, nil)

	// Проверяем, что мапа состояний пуста
	if len(sm.userStates) != 0 {
		t.Errorf("expected 0 user states, got %d", len(sm.userStates))
	}
}

func TestStateMachine_CompleteScenario_NilState(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CompleteScenario panicked with nil state: %v", r)
		}
	}()

	sm := NewStateMachine()
	chatId := int64(1000)
	sm.RegisterScenario("register", []Step{})
	sm.StartScenario(chatId, "register")

	// Завершаем с nil состоянием — не должно паниковать
	sm.CompleteScenario(chatId, nil)

	state := sm.FindUserState(chatId)
	if state != nil {
		t.Errorf("expected state to be nil after complete")
	}
}

func TestStateMachine_MultipleUsers(t *testing.T) {
	sm := NewStateMachine()
	chatId1, chatId2 := int64(1000), int64(2000)

	sm.RegisterScenario("register", []Step{})
	sm.RegisterScenario("login", []Step{})

	sm.StartScenario(chatId1, "register")
	sm.StartScenario(chatId2, "login")

	state1 := sm.FindUserState(chatId1)
	state2 := sm.FindUserState(chatId2)

	if state1 == nil {
		t.Fatal("state1 should not be nil")
	}
	if state2 == nil {
		t.Fatal("state2 should not be nil")
	}

	if state1.CurrentCommand == state2.CurrentCommand {
		t.Fatalf("users should have different commands: both have '%s'", state1.CurrentCommand)
	}

	if state1.CurrentCommand != "register" {
		t.Errorf("expected state1 command 'register', got '%s'", state1.CurrentCommand)
	}

	if state2.CurrentCommand != "login" {
		t.Errorf("expected state2 command 'login', got '%s'", state2.CurrentCommand)
	}

	// Проверяем, что состояния независимы
	state1.Data["username"] = "user1"
	state2.Data["username"] = "user2"

	if state1.Data["username"] != "user1" {
		t.Errorf("expected state1 username 'user1', got '%s'", state1.Data["username"])
	}
	if state2.Data["username"] != "user2" {
		t.Errorf("expected state2 username 'user2', got '%s'", state2.Data["username"])
	}
}

func TestStateMachine_DataPersistence(t *testing.T) {
	sm := NewStateMachine()
	chatId := int64(1000)

	sm.RegisterScenario("register", []Step{})
	sm.StartScenario(chatId, "register")

	// Записываем данные
	state := sm.FindUserState(chatId)
	state.Data["username"] = "testuser"
	state.Data["email"] = "test@example.com"
	state.Data["step"] = "1"

	// Читаем данные повторно (симуляция следующего шага)
	state = sm.FindUserState(chatId)

	if state.Data["username"] != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", state.Data["username"])
	}
	if state.Data["email"] != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", state.Data["email"])
	}
	if state.Data["step"] != "1" {
		t.Errorf("expected step '1', got '%s'", state.Data["step"])
	}
}
