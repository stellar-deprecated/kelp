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
			skipIfGoosNotAllowed(t, k.runGoos)

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
			ospath3, e := ospath1.JoinRelPath(makeOSPath("subfolder", "subfolder", true))
			if !assert.NoError(t, e) {
				return
			}
			// after joining a relativce path we end up with a non-relative path
			if !assert.Equal(t, ospath1.IsRelative(), ospath3.IsRelative()) {
				return
			}
			if !assert.Equal(t, k.wantFinalNative, ospath3.Native()) {
				return
			}
			if !assert.Equal(t, ospath1.Unix()+"/subfolder", ospath3.Unix()) {
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

func TestConvertBetweenNativeUnixPaths(t *testing.T) {
	testCases := []struct {
		name         string
		runGoos      []string
		baseNative   string
		targetNative string
		baseUnix     string
		targetUnix   string
	}{
		{
			name:         "unix_forward",
			runGoos:      []string{"linux", "darwin"},
			baseNative:   "/Users/a/test",
			targetNative: "/Users/a/test/b",
			baseUnix:     "/Users/a/test",
			targetUnix:   "/Users/a/test/b",
		}, {
			name:         "unix_backward",
			runGoos:      []string{"linux", "darwin"},
			baseNative:   "/Users/a/test",
			targetNative: "/Users/a",
			baseUnix:     "/Users/a/test",
			targetUnix:   "/Users/a",
		}, {
			name:         "windows_forward",
			runGoos:      []string{"windows"},
			baseNative:   "C:\\Users\\a\\test",
			targetNative: "C:\\Users\\a\\test\\b",
			baseUnix:     "/Users/a/test",
			targetUnix:   "/Users/a/test/b",
		}, {
			name:         "windows_backward",
			runGoos:      []string{"windows"},
			baseNative:   "C:\\Users\\a\\test",
			targetNative: "C:\\Users\\a",
			baseUnix:     "/Users/a/test",
			targetUnix:   "/Users/a",
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			// early exit if running on a disallowed platform to avoid false negatives
			skipIfGoosNotAllowed(t, k.runGoos)

			targetUnix, e := convertNativePathToUnix(k.baseNative, k.targetNative, k.baseUnix)
			if !assert.NoError(t, e) {
				return
			}
			if !assert.Equal(t, k.targetUnix, targetUnix) {
				return
			}

			targetNative, e := convertUnixPathToNative(k.baseUnix, k.targetUnix, k.baseNative)
			if !assert.NoError(t, e) {
				return
			}
			if !assert.Equal(t, k.targetNative, targetNative) {
				return
			}
		})
	}
}

func TestMakeBetweenNativeUnixPaths(t *testing.T) {
	testCases := []struct {
		name             string
		runGoos          []string
		basePathUnix     string
		targetPathUnix   string
		basePathNative   string
		targetPathNative string
	}{
		{
			name:             "unix_forward",
			runGoos:          []string{"linux", "darwin"},
			basePathUnix:     "/mnt/c/testfolder",
			targetPathUnix:   "/mnt/c/testfolder/a",
			basePathNative:   "/mnt/c/testfolder",
			targetPathNative: "/mnt/c/testfolder/a",
		}, {
			name:             "unix_backward",
			runGoos:          []string{"linux", "darwin"},
			basePathUnix:     "/mnt/c/testfolder/subfolder",
			targetPathUnix:   "/mnt/c/a",
			basePathNative:   "/mnt/c/testfolder/subfolder",
			targetPathNative: "/mnt/c/a",
		}, {
			name:             "windows_forward",
			runGoos:          []string{"windows"},
			basePathUnix:     "/mnt/c/testfolder",
			targetPathUnix:   "/mnt/c/testfolder/a",
			basePathNative:   "C:\\testfolder",
			targetPathNative: "C:\\testfolder\\a",
		}, {
			name:             "windows_backward",
			runGoos:          []string{"windows"},
			basePathUnix:     "/mnt/c/testfolder/subfolder",
			targetPathUnix:   "/mnt/c/a",
			basePathNative:   "C:\\testfolder\\subfolder",
			targetPathNative: "C:\\a",
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			// early exit if running on a disallowed platform to avoid false negatives
			skipIfGoosNotAllowed(t, k.runGoos)

			basepath := makeOSPath(k.basePathNative, k.basePathUnix, false)

			nativePath, e := basepath.MakeFromUnixPath(k.targetPathUnix)
			if !assert.NoError(t, e) {
				return
			}
			if !assert.Equal(t, k.targetPathUnix, nativePath.Unix()) {
				return
			}
			if !assert.Equal(t, k.targetPathNative, nativePath.Native()) {
				return
			}
			if !assert.Equal(t, false, nativePath.IsRelative()) {
				return
			}

			unixPath, e := basepath.MakeFromNativePath(k.targetPathNative)
			if !assert.NoError(t, e) {
				return
			}
			if !assert.Equal(t, k.targetPathUnix, unixPath.Unix()) {
				return
			}
			if !assert.Equal(t, k.targetPathNative, unixPath.Native()) {
				return
			}
			if !assert.Equal(t, false, unixPath.IsRelative()) {
				return
			}
		})
	}
}

func skipIfGoosNotAllowed(t *testing.T, runGoos []string) {
	for _, allowedGoos := range runGoos {
		if runtime.GOOS == allowedGoos {
			return
		}
	}
	t.Skipf("allowed GOOS values = %v but runtime.GOOS = %s", runGoos, runtime.GOOS)
}
