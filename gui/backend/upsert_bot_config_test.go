package backend

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stellar/kelp/gui/model2"
	"github.com/stellar/kelp/support/kelpos"
	"github.com/stretchr/testify/assert"
)

func TestIsBotNameValid(t *testing.T) {
	// TODO: Define tables.
	testCases := []struct {
		name      string
		botName   string
		strategy  string
		wantValid bool
		// TODO: Define test struct.
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
		// TODO: Define each test case.
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			// errors would be thrown by the `regex` package, not due to kelp logic
			gotValid, gotErr := isBotNameValid(k.botName)
			assert.NoError(t, gotErr)
			assert.Equal(t, k.wantValid, gotValid)
			// early return on unsupported bot names, so we do not write an invalid filename
			if !k.wantValid {
				return
			}

			// confirm that we can write and delete an empty file to the bot's filepath
			filenamePair := model2.GetBotFilenames(k.botName, "buysell")
			fileBase, e := kelpos.MakeOsPathBase()
			filePath := fileBase.Join(filenamePair.Trader)
			assert.NoError(t, e)
			e = ioutil.WriteFile(filePath.Native(), []byte{}, 0644)
			assert.NoError(t, e)

			e = os.Remove(filePath.Native())
			assert.NoError(t, e)
		})
	}
}
