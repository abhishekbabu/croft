package provider

import (
	"encoding/json"
	"testing"

	"github.com/abhishekbabu/croft/internal/sh"
	"github.com/stretchr/testify/require"
)

func TestNoneMultiplexerIsInert(t *testing.T) {
	require.NoError(t, NoneMultiplexer{}.CreateSession("x", "/tmp", nil))
	require.False(t, NoneMultiplexer{}.Managed(), "the none multiplexer hosts nothing")
	require.False(t, NoneMultiplexer{}.HasWindow("x", "dev"))
}

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

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"plain":      "'plain'",
		"with space": "'with space'",
		"it's":       `'it'\''s'`,
		"":           "''",
	}
	for in, want := range cases {
		require.Equal(t, want, shellQuote(in), "shellQuote(%q)", in)
	}
}

func TestCmuxCommandLine(t *testing.T) {
	got := cmuxCommandLine(
		map[string]string{"API_PORT": "3000", "CROFT_SLUG": "feat"},
		[]string{"claude", "--session-id", "abc"},
	)
	require.Equal(t, "env 'API_PORT=3000' 'CROFT_SLUG=feat' 'claude' '--session-id' 'abc'", got)
}

func TestParseSurfaceSplit(t *testing.T) {
	id, err := parseSurfaceSplit(`{"surface_id":"S-1"}`)
	require.NoError(t, err)
	require.Equal(t, "S-1", id, "flat shape")

	id, err = parseSurfaceSplit(`{"result":{"surface_id":"S-2"}}`)
	require.NoError(t, err)
	require.Equal(t, "S-2", id, "nested shape")

	_, err = parseSurfaceSplit(`{"ok":true}`)
	require.Error(t, err, "response with no surface id should fail")

	_, err = parseSurfaceSplit(`not json`)
	require.Error(t, err, "invalid JSON should fail")
}

func TestCmuxWorkspaceHasSurface(t *testing.T) {
	raw := `{"windows":[{"workspaces":[{"id":"WS","title":"demo-feat",
	  "panes":[{"surfaces":[{"id":"S-0"}]},{"surfaces":[{"id":"S-1"}]}]}]}]}`
	var tree cmuxTree
	require.NoError(t, json.Unmarshal([]byte(raw), &tree))

	ws := tree.Windows[0].Workspaces[0]
	require.True(t, ws.hasSurface("S-1"), "hasSurface should find S-1 across panes")
	require.False(t, ws.hasSurface("S-9"), "hasSurface should not find a missing surface")
}

func TestCmuxWindowMap(t *testing.T) {
	c := NewCmuxMultiplexer("cmux", t.TempDir())

	_, ok := c.windowSurface("ws", "dev")
	require.False(t, ok, "empty map should not resolve a window")

	require.NoError(t, c.trackWindow("ws", "dev", "S-1"))
	require.NoError(t, c.trackWindow("ws", "agent", "S-2"))

	// A fresh instance must see the persisted map.
	c2 := NewCmuxMultiplexer("cmux", c.stateDir)
	id, ok := c2.windowSurface("ws", "dev")
	require.True(t, ok)
	require.Equal(t, "S-1", id)

	require.NoError(t, c2.forgetWorkspace("ws"))
	_, ok = c2.windowSurface("ws", "agent")
	require.False(t, ok, "forgetWorkspace should drop every window")
}

func TestCmuxRequiresSurface(t *testing.T) {
	c := NewCmuxMultiplexer("cmux", t.TempDir())
	c.surfaceID = "" // simulate running outside cmux
	require.Error(t, c.CreateSession("ws", "/tmp", nil), "CreateSession outside cmux should fail")
	require.Error(t, c.RunWindow("ws", "dev", "/tmp", nil, []string{"true"}), "RunWindow outside cmux should fail")
}
