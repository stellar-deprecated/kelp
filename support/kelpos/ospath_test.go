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

			ospath1 := &OSPath{
				native: k.basePathNative,
				unix:   "/mnt/c/testfolder",
			}

			ospath2 := ospath1.NewPathByAppending("subfolder")
			if !assert.Equal(t, k.wantFinalNative, ospath2.Native()) {
				return
			}
			if !assert.Equal(t, ospath1.Unix()+"/subfolder", ospath2.Unix()) {
				return
			}
		})
	}
}
