package command

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunHooksEmpty(t *testing.T) {
	var out strings.Builder
	require.NoError(t, runHooks("pre", nil, t.TempDir(), nil, &out))
	require.Empty(t, out.String(), "no hooks means no output")
}

func TestRunHooksSuccess(t *testing.T) {
	var out strings.Builder
	err := runHooks("pre", []string{"echo one", "echo two"}, t.TempDir(), nil, &out)
	require.NoError(t, err)
	require.Contains(t, out.String(), "one")
	require.Contains(t, out.String(), "two")
}

func TestRunHooksStopsOnFailure(t *testing.T) {
	var out strings.Builder
	err := runHooks("pre", []string{"echo first", "exit 1", "echo third"}, t.TempDir(), nil, &out)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pre hook failed")
	require.Contains(t, out.String(), "first")
	require.NotContains(t, out.String(), "third", "a failed hook must stop the sequence")
}

func TestRunHooksExportsEnv(t *testing.T) {
	var out strings.Builder
	err := runHooks("pre", []string{"echo $CROFT_HOOK_VAR"}, t.TempDir(),
		map[string]string{"CROFT_HOOK_VAR": "exported"}, &out)
	require.NoError(t, err)
	require.Contains(t, out.String(), "exported", "env should be visible to the hook command")
}

func TestRunHooksRunsInDir(t *testing.T) {
	dir := t.TempDir()
	var out strings.Builder
	require.NoError(t, runHooks("pre", []string{"pwd"}, dir, nil, &out))
	require.Contains(t, out.String(), dir, "the hook should run in the given directory")
}
