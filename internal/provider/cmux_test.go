package provider

import (
	"encoding/json"
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
	"github.com/stretchr/testify/require"
)

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

func TestNewSelectsCmux(t *testing.T) {
	t.Setenv("CMUX_SURFACE_ID", "test-surface")
	set, err := New(config.ProvidersSection{Multiplexer: config.MultiplexerCmux}, config.MachineConfig{}, t.TempDir())
	require.NoError(t, err)
	require.IsType(t, &CmuxMultiplexer{}, set.Multiplexer)
}

func TestNewRejectsCmuxOutsideCmux(t *testing.T) {
	t.Setenv("CMUX_SURFACE_ID", "")
	_, err := New(config.ProvidersSection{Multiplexer: config.MultiplexerCmux}, config.MachineConfig{}, t.TempDir())
	require.Error(t, err, "multiplexer=cmux outside a cmux terminal should fail")
}
