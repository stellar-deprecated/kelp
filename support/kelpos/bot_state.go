package kelpos

import "fmt"

type botState uint8

const (
	botStateInitializing botState = iota
	botStateStopped
	botStateRunning
	botStateStopping
)

// String impl
func (bs botState) String() string {
	return []string{
		"initializing",
		"stopped",
		"running",
		"stopping",
	}[bs]
}

// InitState is the first state of the bot
func InitState() botState {
	return botStateInitializing
}

// nextState produces the next state of the bot
func nextState(bs botState) (botState, error) {
	switch bs {
	case botStateInitializing:
		return botStateStopped, nil
	case botStateStopped:
		return botStateRunning, nil
	case botStateRunning:
		return botStateStopping, nil
	case botStateStopping:
		return botStateStopped, nil
	default:
		return botStateInitializing, fmt.Errorf("botState does not have a nextState: %s", bs.String())
	}
}
