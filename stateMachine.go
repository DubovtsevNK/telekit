package telekit

import (
	"go.uber.org/zap"
)

func NewStateMachine() *StateMachine {
	return &StateMachine{
		userStates: make(map[int64]*UserState),
		scenarios:  make(map[string]Scenario),
	}
}

func (bsm *StateMachine) RegisterScenario(command string, steps []Step) {
	bsm.scenarios[command] = Scenario{Steps: steps}
	log.Info("StateMachine Scenario registered",
		zap.String("command", command))
}

func (bsm *StateMachine) StartScenario(chatID int64, command string) {
	bsm.userStates[chatID] = &UserState{
		CurrentCommand: command,
		Step:           0,
		Data:           make(map[string]string),
	}
	log.Info("StateMachine start scenario",
		zap.Int64("ChatID", chatID),
		zap.String("command", command))
}

func (bsm *StateMachine) FindUserState(chatID int64) *UserState {
	return bsm.userStates[chatID]
}

func (bsm *StateMachine) GetCountScenarioStepInComand(comand string) int {
	return len(bsm.scenarios[comand].Steps)
}

func (bsm *StateMachine) CompleteScenario(chatID int64, state *UserState) {
	delete(bsm.userStates, chatID)
	if state != nil {
		log.Info("StateMachine complete scenario",
			zap.Int64("ChatID", chatID),
			zap.String("command", state.CurrentCommand))
	} else {
		log.Info("StateMachine complete scenario",
			zap.Int64("ChatID", chatID))
	}
}
