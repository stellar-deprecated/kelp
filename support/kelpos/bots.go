package kelpos

import (
	"fmt"
	"log"
	"strings"

	"github.com/stellar/kelp/gui/model"
)

// RegisterBot registers a new bot, returning an error if one already exists with the same name
func (kos *KelpOS) RegisterBot(bot *model.Bot) error {
	return kos.RegisterBotWithState(bot, InitState())
}

// SafeUnregisterBot unregister a bot without any errors
func (kos *KelpOS) SafeUnregisterBot(botName string) {
	kos.botLock.Lock()
	defer kos.botLock.Unlock()

	if _, exists := kos.bots[botName]; exists {
		delete(kos.bots, botName)
		log.Printf("unregistered bot with name '%s'", botName)
	} else {
		log.Printf("no bot registered with name '%s'", botName)
	}
}

// RegisterBotWithState registers a new bot with a given state, returning an error if one already exists with the same name
func (kos *KelpOS) RegisterBotWithState(bot *model.Bot, state BotState) error {
	return kos.registerBotWithState(bot, state, false)
}

// RegisterBotWithStateUpsert registers a new bot with a given state, it always registers the bot even if it is already registered, never returning an error
func (kos *KelpOS) RegisterBotWithStateUpsert(bot *model.Bot, state BotState) {
	_ = kos.registerBotWithState(bot, state, true)
}

// registerBotWithState registers a new bot with a given state, returning an error if one already exists with the same name
func (kos *KelpOS) registerBotWithState(bot *model.Bot, state BotState, forceRegister bool) error {
	kos.botLock.Lock()
	defer kos.botLock.Unlock()

	_, exists := kos.bots[bot.Name]
	if exists {
		if !forceRegister {
			return fmt.Errorf("bot '%s' already registered", bot.Name)
		}
		log.Printf("bot '%s' already registered, but re-registering with state '%s' because forceRegister was set", bot.Name, state)
	}

	kos.bots[bot.Name] = &BotInstance{
		Bot:   bot,
		State: state,
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

// QueryBotState checks to see if the bot is actually running and returns the state accordingly
func (kos *KelpOS) QueryBotState(botName string) (BotState, error) {
	if bi, e := kos.GetBot(botName); e == nil {
		// read initializing state from memory because it's hard to figure that out from the logic below
		if bi.State == BotStateInitializing {
			return bi.State, nil
		}
	}

	prefix := getBotNamePrefix(botName)
	command := fmt.Sprintf("ps aux | grep trade | grep %s | grep -v grep", prefix)
	outputBytes, e := kos.Blocking("query_bot_state", command)
	if e != nil {
		if strings.Contains(e.Error(), "exit status 1") {
			return BotStateStopped, nil
		}
		return InitState(), fmt.Errorf("error querying bot state using command '%s': %s", command, e)
	}
	output := strings.TrimSpace(string(outputBytes))

	if strings.Contains(output, "delete") {
		return BotStateStopping, nil
	}
	return BotStateRunning, nil
}

// RegisteredBots returns the list of registered bots
func (kos *KelpOS) RegisteredBots() []string {
	list := []string{}
	for k, _ := range kos.bots {
		list = append(list, k)
	}
	return list
}

// getBotNamePrefix returns the general prefix for filenames associated with a botName
func getBotNamePrefix(botName string) string {
	return strings.ToLower(strings.Replace(botName, " ", "_", -1))
}
