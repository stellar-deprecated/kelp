package kelpos

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/stellar/kelp/support/utils"
)

// OSPath encapsulates the pair of the native path (i.e. windows or unix) and the unix path
// this allows certain commands which are unix-specific to have access to the path instead of running transformations
type OSPath struct {
	native string
	unix   string
	isRel  bool
}

// String() is the Stringer method which is unsupprted
func (o *OSPath) String() string {
	log.Fatalf("String method is unsupported because the usage is ambiguous for this struct, use .Unix(), .Native(), or .AsString() instead, trace:\n%s", string(debug.Stack()))
	return ""
}

// AsString produces a string representation and we intentionally don't use the Stringer API because this can mistakenly
// be used in place of a string path which will produce hidden runtime errors which is dangerous
func (o *OSPath) AsString() string {
	return fmt.Sprintf("OSPath[native=%s, unix=%s, isRel=%v]", o.native, o.unix, o.isRel)
}

// makeOSPath is an internal helper that enforced always using toUnixFilepath on the unix path
func makeOSPath(native string, unix string, isRel bool) *OSPath {
	return &OSPath{
		native: filepath.Clean(native),
		unix:   toUnixFilepath(filepath.Clean(unix)),
		isRel:  isRel,
	}
}

// MakeOsPathBase is a factory method for the OSPath struct based on the current binary's directory
func MakeOsPathBase() (*OSPath, error) {
	binaryDirectoryNative, e := getBinaryDirectoryNative()
	if e != nil {
		return nil, fmt.Errorf("could not get binary directory: %s", e)
	}

	binaryDirectoryUnix, e := getBinaryDirectoryUnix(binaryDirectoryNative)
	if e != nil {
		return nil, fmt.Errorf("could not get binary directory unix: %s", e)
	}

	// use the binary directory for the base path since UI executables will be run from $PWD whereas the binary can be located
	// in a different path (for example $PWD/Kelp.app/Contents/MacOS/ for darwin)
	ospath := makeOSPath(binaryDirectoryNative, binaryDirectoryUnix, false)

	if filepath.Base(ospath.Native()) != filepath.Base(ospath.Unix()) {
		errorStr := fmt.Sprintf("ran from directory (%s) but need to run from the same directory as the location of the binary (%s), cd over to the location of the binary\n", ospath.Unix(), ospath.Native())
		utils.PrintErrorHintf(errorStr)
		return nil, fmt.Errorf(errorStr)
	}

	return ospath, nil
}

// Native returns the native representation of the path as a string
func (o *OSPath) Native() string {
	return o.native
}

// Unix returns the unix representation of the path as a string
func (o *OSPath) Unix() string {
	return o.unix
}

// IsRelative returns true if this is a relative path, otherwise false
func (o *OSPath) IsRelative() bool {
	return o.isRel
}

// Join makes a new OSPath struct by modifying the internal path representations together
func (o *OSPath) Join(elem ...string) *OSPath {
	nativePaths := []string{o.native}
	nativePaths = append(nativePaths, elem...)

	unixPaths := []string{o.unix}
	unixPaths = append(unixPaths, elem...)

	return makeOSPath(
		filepath.Join(nativePaths...),
		filepath.Join(unixPaths...),
		o.isRel,
	)
}

// JoinRelPath makes a new OSPath struct by modifying the internal path representations together
func (o *OSPath) JoinRelPath(relPaths ...*OSPath) (*OSPath, error) {
	elems := []string{}
	for _, path := range relPaths {
		if !path.IsRelative() {
			return nil, fmt.Errorf("paths need to be relative but found a non-relative path %s", path.AsString())
		}
		elems = append(elems, path.Native())
	}
	return o.Join(elems...), nil
}

// RelFromBase returns a *OSPath that is relative to the basepath
func (o *OSPath) RelFromBase() (*OSPath, error) {
	basepath, e := MakeOsPathBase()
	if e != nil {
		return nil, fmt.Errorf("unable to fetch ospathbase: %s", e)
	}

	return o.RelFromPath(basepath)
}

// RelFromPath returns a *OSPath that is relative from the provided path
func (o *OSPath) RelFromPath(basepath *OSPath) (*OSPath, error) {
	newRelNative, e := filepath.Rel(basepath.Native(), o.Native())
	if e != nil {
		return nil, fmt.Errorf("unable to make relative native path from basepath: %s", e)
	}

	newRelUnix, e := filepath.Rel(basepath.Unix(), o.Unix())
	if e != nil {
		return nil, fmt.Errorf("unable to make relative unix path from basepath: %s", e)
	}

	// set this to be a relative path
	return makeOSPath(newRelNative, newRelUnix, true), nil
}

// getBinaryDirectoryUnix takes the workingDirUnix and adds to it the relative path from
// workingDirNative to binaryDirectoryNative in unix form.
func getBinaryDirectoryUnix(binaryDirectoryNative string) (string, error) {
	wdUnix, e := getWorkingDirUnix()
	if e != nil {
		return "", fmt.Errorf("could not fetch working directory unix: %s", e)
	}

	wdNative, e := os.Getwd()
	if e != nil {
		return "", fmt.Errorf("could not fetch working directory native: %s", e)
	}

	// on windows the native format is not unix-based so we need to convert from C:\ vs /mnt/c notation. This method will
	// standardize conversions across OSes but results in a nop for darwin/unix since they use unix paths as their native format
	return convertNativePathToUnix(wdNative, binaryDirectoryNative, wdUnix)
}

func convertNativePathToUnix(baseNative string, targetNative string, baseUnix string) (string, error) {
	relBaseToTarget, e := filepath.Rel(baseNative, targetNative)
	if e != nil {
		return "", fmt.Errorf("could not fetch relative path from baseNative (%s) to targetNative (%s): %s", baseNative, targetNative, e)
	}

	return toUnixFilepath(filepath.Join(baseUnix, toUnixFilepath(relBaseToTarget))), nil
}

func getWorkingDirUnix() (string, error) {
	outputBytes, e := exec.Command("bash", "-c", "pwd").Output()
	if e != nil {
		return "", fmt.Errorf("could not run raw command 'bash -c pwd': %s", e)
	}
	return strings.TrimSpace(string(outputBytes)), nil
}

func getBinaryDirectoryNative() (string, error) {
	return filepath.Abs(filepath.Dir(os.Args[0]))
}

func toUnixFilepath(path string) string {
	return filepath.ToSlash(path)
}
