package kelpos

import "fmt"

type BotState uint8

const (
	BotStateInitializing BotState = iota
	BotStateStopped
	BotStateRunning
	BotStateStopping
)

// String impl
func (bs BotState) String() string {
	return []string{
		"initializing",
		"stopped",
		"running",
		"stopping",
	}[bs]
}

// InitState is the first state of the bot
func InitState() BotState {
	return BotStateInitializing
}

// nextState produces the next state of the bot
func nextState(bs BotState) (BotState, error) {
	switch bs {
	case BotStateInitializing:
		return BotStateStopped, nil
	case BotStateStopped:
		return BotStateRunning, nil
	case BotStateRunning:
		return BotStateStopping, nil
	case BotStateStopping:
		return BotStateStopped, nil
	default:
		return BotStateInitializing, fmt.Errorf("botState does not have a nextState: %s", bs.String())
	}
}
