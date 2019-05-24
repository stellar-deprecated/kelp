package kelpos

import (
	"fmt"
	"log"

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
		Bot:   bot,
		State: InitState(),
	}
	return nil
}

// AdvanceBotState advances the state of the given bot atomically, ensuring the bot is currently at the expected state
func (kos *KelpOS) AdvanceBotState(botName string, expectedCurrentState BotState) error {
	kos.botLock.Lock()
	defer kos.botLock.Unlock()

	b, exists := kos.bots[botName]
	if !exists {
		return fmt.Errorf("bot '%s' is not registered", botName)
	}

	if b.State != expectedCurrentState {
		return fmt.Errorf("state of bot '%s' was not as expected (%s): %s", botName, expectedCurrentState, b.State)
	}

	ns, e := nextState(b.State)
	if e != nil {
		return fmt.Errorf("error while advancing bot state for '%s': %s", botName, e)
	}

	b.State = ns
	log.Printf("advanced bot state for bot '%s' to %s\n", botName, ns)

	return nil
}

// GetBot fetches the bot state for the given name
func (kos *KelpOS) GetBot(botName string) (*BotInstance, error) {
	kos.botLock.Lock()
	defer kos.botLock.Unlock()

	b, exists := kos.bots[botName]
	if !exists {
		return b, fmt.Errorf("bot '%s' does not exist", botName)
	}
	return b, nil
}
