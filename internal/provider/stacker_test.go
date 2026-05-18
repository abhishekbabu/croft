package provider

import (
	"reflect"
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
)

func TestParseStackBranches(t *testing.T) {
	// Mimics `gt log short -s` with ANSI codes and tree-drawing characters.
	out := "\x1b[32m◯ roa-3-top\x1b[0m\n" +
		"│\n" +
		"\x1b[33m◯ roa-2-mid\x1b[0m\n" +
		"│\n" +
		"◉ roa-1-base\n" +
		"│\n" +
		"◯ main\n"
	got := parseStackBranches(out)
	want := []string{"roa-3-top", "roa-2-mid", "roa-1-base"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseStackBranches = %v, want %v", got, want)
	}
}

func TestParsePRStates(t *testing.T) {
	data := []byte(`[
		{"headRefName":"roa-1-base","state":"MERGED"},
		{"headRefName":"roa-2-mid","state":"OPEN"}
	]`)
	got := parsePRStates(data)
	if got["roa-1-base"] != "MERGED" || got["roa-2-mid"] != "OPEN" {
		t.Errorf("parsePRStates = %v", got)
	}
	if len(parsePRStates([]byte("not json"))) != 0 {
		t.Error("invalid JSON should yield an empty map")
	}
}

func TestNewSelectsStackerAndRouter(t *testing.T) {
	set, err := New(config.ProvidersSection{
		Router:       "portless",
		Stacker:      "graphite",
		Coordination: "basic",
	}, config.MachineConfig{}, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := set.Router.(*PortlessRouter); !ok {
		t.Errorf("router = %T, want *PortlessRouter", set.Router)
	}
	if _, ok := set.Stacker.(*GraphiteStacker); !ok {
		t.Errorf("stacker = %T, want *GraphiteStacker", set.Stacker)
	}
}
