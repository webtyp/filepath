//go:build !wasm

package filepath

import (
	"os"
)

// detectRoot returns the base path using os.Getwd().
func detectRoot() string {
	if wd, err := os.Getwd(); err == nil {
		cleaned, _ := pathClean(wd)
		return cleaned
	}
	return ""
}

// detectHome returns the user's home directory.
func detectHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	cleaned, _ := pathClean(home)
	return cleaned
}
