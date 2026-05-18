package provider

import (
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
	"github.com/stretchr/testify/require"
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
	require.Equal(t, []string{"roa-3-top", "roa-2-mid", "roa-1-base"}, parseStackBranches(out))
}

func TestParsePRStates(t *testing.T) {
	data := []byte(`[
		{"headRefName":"roa-1-base","state":"MERGED"},
		{"headRefName":"roa-2-mid","state":"OPEN"}
	]`)
	got := parsePRStates(data)
	require.Equal(t, "MERGED", got["roa-1-base"])
	require.Equal(t, "OPEN", got["roa-2-mid"])
	require.Empty(t, parsePRStates([]byte("not json")), "invalid JSON should yield an empty map")
}

func TestNewSelectsStackerAndRouter(t *testing.T) {
	set, err := New(config.ProvidersSection{
		Router:       config.RouterPortless,
		Stacker:      config.StackerGraphite,
		Coordination: config.CoordinationBasic,
	}, config.MachineConfig{}, "")
	require.NoError(t, err)
	require.IsType(t, &PortlessRouter{}, set.Router)
	require.IsType(t, &GraphiteStacker{}, set.Stacker)
}
