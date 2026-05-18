package provider

import (
	"encoding/json"
	"testing"

	"github.com/abhishekbabu/croft/internal/testutil"
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

// --- shell-out tests against a fake cmux ---

// fakeCmux builds a stand-in cmux that answers the subcommands croft issues:
// `tree` and `identify` return the templated JSON; `rpc`/`read-screen` return
// fixed responses; everything else just succeeds.
func fakeCmux(t *testing.T, tree, identify string) string {
	t.Helper()
	script := `case "$1" in
  --json)      printf '%s' '` + tree + `' ;;
  identify)    printf '%s' '` + identify + `' ;;
  rpc)         echo '{"surface_id":"S-new"}' ;;
  read-screen) printf 'pane text' ;;
esac
exit 0`
	return testutil.FakeBin(t, "cmux", script)
}

// demoFeatTree is a cmux tree with the "demo-feat" workspace holding surface S-1.
const demoFeatTree = `{"windows":[{"workspaces":[{"id":"WS-1","title":"demo-feat","panes":[{"surfaces":[{"id":"S-1"}]}]}]}]}`

func TestCmuxManaged(t *testing.T) {
	require.True(t, NewCmuxMultiplexer("cmux", t.TempDir()).Managed())
}

func TestCmuxCreateSession(t *testing.T) {
	t.Setenv("CMUX_SURFACE_ID", "S-self")
	c := NewCmuxMultiplexer(fakeCmux(t, `{"windows":[]}`, ""), t.TempDir())
	require.NoError(t, c.CreateSession("demo-feat", "/tmp", nil))
}

func TestCmuxAttach(t *testing.T) {
	t.Setenv("CMUX_SURFACE_ID", "S-self")
	c := NewCmuxMultiplexer(fakeCmux(t, demoFeatTree, ""), t.TempDir())
	require.NoError(t, c.Attach("demo-feat"))
	require.Error(t, c.Attach("nope"), "Attach to an unknown workspace should fail")
}

func TestCmuxKill(t *testing.T) {
	t.Setenv("CMUX_SURFACE_ID", "S-self")
	c := NewCmuxMultiplexer(fakeCmux(t, demoFeatTree, ""), t.TempDir())
	require.NoError(t, c.Kill("demo-feat"))
	require.NoError(t, c.Kill("absent"), "Kill of an absent workspace is a no-op")
}

func TestCmuxHasWindowAndCapturePane(t *testing.T) {
	t.Setenv("CMUX_SURFACE_ID", "S-self")
	c := NewCmuxMultiplexer(fakeCmux(t, demoFeatTree, ""), t.TempDir())

	require.False(t, c.HasWindow("demo-feat", "dev"), "no window tracked yet")
	require.NoError(t, c.trackWindow("demo-feat", "dev", "S-1"))
	require.True(t, c.HasWindow("demo-feat", "dev"), "tracked surface is present in the tree")

	out, err := c.CapturePane("demo-feat", "dev", 10)
	require.NoError(t, err)
	require.Equal(t, "pane text", out)

	_, err = c.CapturePane("demo-feat", "missing", 10)
	require.Error(t, err, "CapturePane of an untracked window should fail")
}

func TestCmuxRunWindow(t *testing.T) {
	t.Setenv("CMUX_SURFACE_ID", "S-self")
	identify := `{"caller":{"surface_ref":"S-self"},"focused":{"surface_ref":"S-self"}}`
	c := NewCmuxMultiplexer(fakeCmux(t, demoFeatTree, identify), t.TempDir())

	require.NoError(t, c.RunWindow("demo-feat", "agent", "/tmp", nil, []string{"true"}))

	id, ok := c.windowSurface("demo-feat", "agent")
	require.True(t, ok, "RunWindow should track the new window")
	require.Equal(t, "S-new", id, "the tracked surface is the one cmux split off")
}
