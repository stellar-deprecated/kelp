package backend

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stellar/kelp/gui/model2"
	"github.com/stretchr/testify/assert"
)

func TestIsBotNameValid(t *testing.T) {
	testCases := []struct {
		name      string
		botName   string
		strategy  string
		wantValid bool
	}{
		{
			name:      "success default",
			botName:   "George the Friendly Octopus",
			wantValid: true,
		},
		{
			name:      "success numeric",
			botName:   "George the 1337 Octopus",
			wantValid: true,
		},
		{
			name:      "success dash",
			botName:   "George-the-Friendly-Octopus",
			wantValid: true,
		},
		{
			name:      "failure <>",
			botName:   "George<>the Friendly Octopus",
			wantValid: false,
		},
		{
			name:      "failure \\/",
			botName:   "George\\/the Friendly Octopus",
			wantValid: false,
		},
		{
			name:      "failure .",
			botName:   "George . the Friendly Octopus",
			wantValid: false,
		},
		{
			name:      "failure &",
			botName:   "George & the Friendly Octopus",
			wantValid: false,
		},
		{
			name:      "failure _",
			botName:   "George _ the Friendly Octopus",
			wantValid: false,
		},
		{
			name:      "failure ()",
			botName:   "George () the Friendly Octopus",
			wantValid: false,
		},
		{
			name:      "failure +",
			botName:   "George + the Friendly Octopus",
			wantValid: false,
		},
		{
			name:      "failure @",
			botName:   "George @ the Friendly Octopus",
			wantValid: false,
		},
	}

	// initialize botname regex
	e := InitBotNameRegex()
	if !assert.NoError(t, e) {
		t.Log("Could not generate botname regex")
		return
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			gotValid := isBotNameValid(k.botName)
			if !assert.Equal(t, k.wantValid, gotValid) {
				return
			}

			// early return on unsupported bot names, so we do not write an invalid filename
			if !k.wantValid {
				return
			}

			// confirm that we can write and delete an empty file to the bot's filepath
			filenamePair := model2.GetBotFilenames(k.botName, "buysell")
			fileBase := os.TempDir()
			filePath := fmt.Sprintf("%s/%s", fileBase, filenamePair.Trader)
			e = ioutil.WriteFile(filePath, []byte{}, 0644)
			if !assert.NoError(t, e) {
				return
			}

			e = os.Remove(filePath)
			assert.NoError(t, e)
		})
	}
}
