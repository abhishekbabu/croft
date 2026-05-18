package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCmuxSurfaceID(t *testing.T) {
	t.Setenv("CMUX_SURFACE_ID", "surface-42")
	require.Equal(t, "surface-42", CmuxSurfaceID())

	t.Setenv("CMUX_SURFACE_ID", "")
	require.Empty(t, CmuxSurfaceID(), "an unset variable yields the empty string")
}

func TestConfigHomeHonorsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/cfg")
	dir, err := ConfigHome()
	require.NoError(t, err)
	require.Equal(t, "/custom/cfg", dir)
}

func TestConfigHomeFallsBackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	dir, err := ConfigHome()
	require.NoError(t, err)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, ".config"), dir)
}

func TestDataHomeHonorsXDG(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/custom/data")
	dir, err := DataHome()
	require.NoError(t, err)
	require.Equal(t, "/custom/data", dir)
}

func TestDataHomeFallsBackToHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	dir, err := DataHome()
	require.NoError(t, err)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, ".local", "share"), dir)
}
