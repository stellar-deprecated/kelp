package kelpos

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSPath(t *testing.T) {
	testCases := []struct {
		runGoos         []string
		basePathNative  string
		wantFinalNative string
	}{
		{
			runGoos:         []string{"linux", "darwin"},
			basePathNative:  "/mnt/c/testfolder",
			wantFinalNative: "/mnt/c/testfolder/subfolder",
		}, {
			runGoos:         []string{"windows"},
			basePathNative:  "C:\\testfolder",
			wantFinalNative: "C:\\testfolder\\subfolder",
		},
	}

	for _, k := range testCases {
		t.Run(k.basePathNative, func(t *testing.T) {
			// early exit if running on a disallowed platform to avoid false negatives
			isValid := false
			for _, allowedGoos := range k.runGoos {
				if runtime.GOOS == allowedGoos {
					isValid = true
					break
				}
			}
			if !isValid {
				return
			}

			ospath1 := makeOSPath(k.basePathNative, "/mnt/c/testfolder", false)
			if !assert.Equal(t, false, ospath1.IsRelative()) {
				return
			}
			ospath2 := ospath1.Join("subfolder")
			if !assert.Equal(t, false, ospath2.IsRelative()) {
				return
			}
			if !assert.Equal(t, k.wantFinalNative, ospath2.Native()) {
				return
			}
			if !assert.Equal(t, ospath1.Unix()+"/subfolder", ospath2.Unix()) {
				return
			}

			rel1, e := ospath2.RelFromPath(ospath1)
			if !assert.NoError(t, e) {
				return
			}
			if !assert.Equal(t, true, rel1.IsRelative()) {
				return
			}
			if !assert.Equal(t, "subfolder", rel1.Unix()) {
				return
			}
		})
	}
}
