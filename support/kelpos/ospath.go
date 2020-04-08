package kelpos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// OSPath encapsulates the pair of the native path (i.e. windows or unix) and the unix path
// this allows certain commands which are unix-specific to have access to the path instead of running transformations
type OSPath struct {
	native string
	unix   string
}

// MakeOsPath is a factory method for the OSPath struct based on the current binary's directory
func MakeOsPath() (*OSPath, error) {
	currentDirUnslashed, e := getCurrentDirUnix()
	if e != nil {
		return nil, fmt.Errorf("could not get current directory: %s", e)
	}
	currentDirUnix := toUnixFilepath(currentDirUnslashed)

	binaryDirectoryNative, e := getBinaryDirectoryNative()
	if e != nil {
		return nil, fmt.Errorf("could not get binary directory: %s", e)
	}

	if filepath.Base(currentDirUnix) != filepath.Base(binaryDirectoryNative) {
		return nil, fmt.Errorf("ran from directory (%s) but need to run from the same directory as the location of the binary (%s), cd over to the location of the binary", currentDirUnix, binaryDirectoryNative)
	}

	return &OSPath{
		native: binaryDirectoryNative,
		unix:   currentDirUnix,
	}, nil
}

// Native returns the native representation of the path as a string
func (o *OSPath) Native() string {
	return o.native
}

// Unix returns the unix representation of the path as a string
func (o *OSPath) Unix() string {
	return o.unix
}

// NewPathByAppending makes a new OSPath struct by modifying the internal path representations together
func (o *OSPath) NewPathByAppending(elem ...string) *OSPath {
	nativePaths := []string{o.native}
	nativePaths = append(nativePaths, elem...)

	unixPaths := []string{o.unix}
	unixPaths = append(unixPaths, elem...)

	return &OSPath{
		native: filepath.Join(nativePaths...),
		unix:   filepath.Join(unixPaths...),
	}
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
