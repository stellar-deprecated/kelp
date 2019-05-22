package kelpos

import (
	"fmt"

	"github.com/stellar/kelp/gui/model"
)

// RegisterBot registers a new bot, returning an error if one already exists with the same name
func (kos *KelpOS) RegisterBot(bot *model.Bot) error {
	kos.botLock.Lock()
	defer kos.botLock.Unlock()

	if _, exists := kos.bots[bot.Name]; exists {
		return fmt.Errorf("bot '%s' already registered", bot.Name)
	}

	kos.bots[bot.Name] = &BotInstance{
		bot:   bot,
		state: InitState(),
	}
	return nil
}

// AdvanceBotState advances the state of the given bot atomically, ensuring the bot is currently at the expected state
func (kos *KelpOS) AdvanceBotState(botName string, expectedCurrentState botState) error {
	kos.botLock.Lock()
	defer kos.botLock.Unlock()

	b, exists := kos.bots[botName]
	if !exists {
		return fmt.Errorf("bot '%s' is not registered", botName)
	}

	if b.state != expectedCurrentState {
		return fmt.Errorf("state of bot '%s' was not as expected (%s): %s", botName, expectedCurrentState, b.state)
	}

	ns, e := nextState(b.state)
	if e != nil {
		return fmt.Errorf("error while advancing bot state for '%s': %s", botName, e)
	}
	b.state = ns
	return nil
}

func (kos *KelpOS) getBot(name string) (*BotInstance, bool) {
	kos.botLock.Lock()
	defer kos.botLock.Unlock()

	b, exists := kos.bots[name]
	return b, exists
}
