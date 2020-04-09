package kelpos

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
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
	currentDirUnslashed, e := getCurrentDirUnix()
	if e != nil {
		return nil, fmt.Errorf("could not get current directory: %s", e)
	}
	binaryDirectoryNative, e := getBinaryDirectoryNative()
	if e != nil {
		return nil, fmt.Errorf("could not get binary directory: %s", e)
	}
	ospath := makeOSPath(binaryDirectoryNative, currentDirUnslashed, false)

	if filepath.Base(ospath.Native()) != filepath.Base(ospath.Unix()) {
		return nil, fmt.Errorf("ran from directory (%s) but need to run from the same directory as the location of the binary (%s), cd over to the location of the binary", ospath.Unix(), ospath.Native())
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

func getCurrentDirUnix() (string, error) {
	kos := GetKelpOS()
	outputBytes, e := kos.Blocking("pwd", "pwd")
	if e != nil {
		return "", fmt.Errorf("could not fetch current directory: %s", e)
	}
	return strings.TrimSpace(string(outputBytes)), nil
}

func getBinaryDirectoryNative() (string, error) {
	return filepath.Abs(filepath.Dir(os.Args[0]))
}

func toUnixFilepath(path string) string {
	return filepath.ToSlash(path)
}
