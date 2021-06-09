package kelpos

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/stellar/kelp/gui/model2"
)

// BotInstance is an instance of a given bot along with the metadata
type BotInstance struct {
	Bot   *model2.Bot
	State BotState
}

// UserBotData represents the Bot registration map and other items related to a given user
type UserBotData struct {
	user    *User
	kos     *KelpOS
	bots    map[string]*BotInstance
	botLock *sync.Mutex
}

// makeUserBotData is a factory method
func makeUserBotData(kos *KelpOS, user *User) *UserBotData {
	return &UserBotData{
		kos:     kos,
		user:    user,
		bots:    map[string]*BotInstance{},
		botLock: &sync.Mutex{},
	}
}

// RegisterBot registers a new bot, returning an error if one already exists with the same name
func (ubd *UserBotData) RegisterBot(bot *model2.Bot) error {
	return ubd.RegisterBotWithState(bot, InitState())
}

// SafeUnregisterBot unregister a bot without any errors
func (ubd *UserBotData) SafeUnregisterBot(botName string) {
	ubd.botLock.Lock()
	defer ubd.botLock.Unlock()

	if _, exists := ubd.bots[botName]; exists {
		delete(ubd.bots, botName)
		log.Printf("unregistered bot with name '%s'", botName)
	} else {
		log.Printf("no bot registered with name '%s'", botName)
	}
}

// RegisterBotWithState registers a new bot with a given state, returning an error if one already exists with the same name
func (ubd *UserBotData) RegisterBotWithState(bot *model2.Bot, state BotState) error {
	return ubd.registerBotWithState(bot, state, false)
}

// RegisterBotWithStateUpsert registers a new bot with a given state, it always registers the bot even if it is already registered, never returning an error
func (ubd *UserBotData) RegisterBotWithStateUpsert(bot *model2.Bot, state BotState) {
	_ = ubd.registerBotWithState(bot, state, true)
}

// registerBotWithState registers a new bot with a given state, returning an error if one already exists with the same name
func (ubd *UserBotData) registerBotWithState(bot *model2.Bot, state BotState, forceRegister bool) error {
	ubd.botLock.Lock()
	defer ubd.botLock.Unlock()

	_, exists := ubd.bots[bot.Name]
	if exists {
		if !forceRegister {
			return fmt.Errorf("bot '%s' already registered", bot.Name)
		}
		log.Printf("bot '%s' already registered, but re-registering with state '%s' because forceRegister was set", bot.Name, state)
	}

	ubd.bots[bot.Name] = &BotInstance{
		Bot:   bot,
		State: state,
	}
	return nil
}

// AdvanceBotState advances the state of the given bot atomically, ensuring the bot is currently at the expected state
func (ubd *UserBotData) AdvanceBotState(botName string, expectedCurrentState BotState) error {
	ubd.botLock.Lock()
	defer ubd.botLock.Unlock()

	b, exists := ubd.bots[botName]
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
func (ubd *UserBotData) GetBot(botName string) (*BotInstance, error) {
	ubd.botLock.Lock()
	defer ubd.botLock.Unlock()

	b, exists := ubd.bots[botName]
	if !exists {
		return b, fmt.Errorf("bot '%s' does not exist", botName)
	}
	return b, nil
}

// QueryBotState checks to see if the bot is actually running and returns the state accordingly
func (ubd *UserBotData) QueryBotState(botName string) (BotState, error) {
	if bi, e := ubd.GetBot(botName); e == nil {
		// read initializing state from memory because it's hard to figure that out from the logic below
		if bi.State == BotStateInitializing {
			return bi.State, nil
		}
	}

	prefix := getBotNamePrefix(botName)
	command := fmt.Sprintf("ps aux | grep trade | grep %s | grep -v grep", prefix)
	outputBytes, e := ubd.kos.Blocking(ubd.user.ID, fmt.Sprintf("query_bot_state: %s", botName), command)
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
func (ubd *UserBotData) RegisteredBots() []string {
	list := []string{}
	for k := range ubd.bots {
		list = append(list, k)
	}
	return list
}

// getBotNamePrefix returns the general prefix for filenames associated with a botName
func getBotNamePrefix(botName string) string {
	return strings.ToLower(strings.Replace(botName, " ", "_", -1))
}
