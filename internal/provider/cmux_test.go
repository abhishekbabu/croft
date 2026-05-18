package provider

import (
	"encoding/json"
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
)

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"plain":      "'plain'",
		"with space": "'with space'",
		"it's":       `'it'\''s'`,
		"":           "''",
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Errorf("shellQuote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCmuxCommandLine(t *testing.T) {
	got := cmuxCommandLine(
		map[string]string{"API_PORT": "3000", "CROFT_SLUG": "feat"},
		[]string{"claude", "--session-id", "abc"},
	)
	want := "env 'API_PORT=3000' 'CROFT_SLUG=feat' 'claude' '--session-id' 'abc'"
	if got != want {
		t.Errorf("cmuxCommandLine =\n  %q\nwant\n  %q", got, want)
	}
}

func TestParseSurfaceSplit(t *testing.T) {
	id, err := parseSurfaceSplit(`{"surface_id":"S-1"}`)
	if err != nil || id != "S-1" {
		t.Errorf("flat shape: id=%q err=%v", id, err)
	}
	id, err = parseSurfaceSplit(`{"result":{"surface_id":"S-2"}}`)
	if err != nil || id != "S-2" {
		t.Errorf("nested shape: id=%q err=%v", id, err)
	}
	if _, err := parseSurfaceSplit(`{"ok":true}`); err == nil {
		t.Error("response with no surface id should error")
	}
	if _, err := parseSurfaceSplit(`not json`); err == nil {
		t.Error("invalid JSON should error")
	}
}

func TestCmuxWorkspaceHasSurface(t *testing.T) {
	raw := `{"windows":[{"workspaces":[{"id":"WS","title":"demo-feat",
	  "panes":[{"surfaces":[{"id":"S-0"}]},{"surfaces":[{"id":"S-1"}]}]}]}]}`
	var tree cmuxTree
	if err := json.Unmarshal([]byte(raw), &tree); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	ws := tree.Windows[0].Workspaces[0]
	if !ws.hasSurface("S-1") {
		t.Error("hasSurface should find S-1 across panes")
	}
	if ws.hasSurface("S-9") {
		t.Error("hasSurface should not find a missing surface")
	}
}

func TestCmuxWindowMap(t *testing.T) {
	c := NewCmuxMultiplexer("cmux", t.TempDir())

	if _, ok := c.windowSurface("ws", "dev"); ok {
		t.Error("empty map should not resolve a window")
	}
	if err := c.trackWindow("ws", "dev", "S-1"); err != nil {
		t.Fatalf("trackWindow: %v", err)
	}
	if err := c.trackWindow("ws", "agent", "S-2"); err != nil {
		t.Fatalf("trackWindow: %v", err)
	}

	// A fresh instance must see the persisted map.
	c2 := NewCmuxMultiplexer("cmux", c.stateDir)
	if id, ok := c2.windowSurface("ws", "dev"); !ok || id != "S-1" {
		t.Errorf("windowSurface(ws,dev) = %q,%v", id, ok)
	}

	if err := c2.forgetWorkspace("ws"); err != nil {
		t.Fatalf("forgetWorkspace: %v", err)
	}
	if _, ok := c2.windowSurface("ws", "agent"); ok {
		t.Error("forgetWorkspace should drop every window")
	}
}

func TestCmuxRequiresSurface(t *testing.T) {
	c := NewCmuxMultiplexer("cmux", t.TempDir())
	c.surfaceID = "" // simulate running outside cmux
	if err := c.CreateSession("ws", "/tmp", nil); err == nil {
		t.Error("CreateSession outside cmux should error")
	}
	if err := c.RunWindow("ws", "dev", "/tmp", nil, []string{"true"}); err == nil {
		t.Error("RunWindow outside cmux should error")
	}
}

func TestNewSelectsCmux(t *testing.T) {
	set, err := New(config.ProvidersSection{Multiplexer: "cmux"}, config.MachineConfig{}, t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := set.Multiplexer.(*CmuxMultiplexer); !ok {
		t.Errorf("multiplexer = %T, want *CmuxMultiplexer", set.Multiplexer)
	}
}
