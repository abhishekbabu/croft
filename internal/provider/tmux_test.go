package provider

import (
	"testing"

	"github.com/abhishekbabu/croft/internal/sh"
	"github.com/stretchr/testify/require"
)

func TestTmuxLifecycle(t *testing.T) {
	if !sh.Look("tmux") {
		t.Skip("tmux not installed")
	}
	tm := NewTmuxMultiplexer("")
	const name = "croft-test-session"
	_ = tm.Kill(name) // clean any leftover
	t.Cleanup(func() { _ = tm.Kill(name) })

	require.NoError(t, tm.CreateSession(name, t.TempDir(), map[string]string{"FOO": "bar"}))
	require.NoError(t, tm.CreateSession(name, t.TempDir(), nil), "CreateSession is idempotent")
	require.True(t, tm.hasSession(name), "session should exist after CreateSession")

	require.NoError(t, tm.Kill(name))
	require.NoError(t, tm.Kill(name), "Kill is idempotent")
}
