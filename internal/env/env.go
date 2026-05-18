// Package env centralizes every environment variable croft reads. Keeping the
// lookups in one place means each variable is documented once and callers
// never depend on a bare os.Getenv with a stringly-typed key.
package env

import (
	"fmt"
	"os"
	"path/filepath"
)

// CmuxSurfaceID returns the cmux surface croft is running in
// ($CMUX_SURFACE_ID), or "" when croft was not launched from a cmux terminal.
func CmuxSurfaceID() string {
	return os.Getenv("CMUX_SURFACE_ID")
}

// ConfigHome returns the base directory for user configuration: $XDG_CONFIG_HOME
// when set, otherwise ~/.config.
func ConfigHome() (string, error) {
	return xdgDir("XDG_CONFIG_HOME", ".config")
}

// DataHome returns the base directory for user data: $XDG_DATA_HOME when set,
// otherwise ~/.local/share.
func DataHome() (string, error) {
	return xdgDir("XDG_DATA_HOME", filepath.Join(".local", "share"))
}

// xdgDir resolves an XDG base directory: the named variable when set, else the
// fallback path relative to the user's home directory.
func xdgDir(envVar, fallback string) (string, error) {
	if dir := os.Getenv(envVar); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}
	return filepath.Join(home, fallback), nil
}
