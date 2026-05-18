package provider

import (
	"testing"

	"github.com/abhishekbabu/croft/internal/testutil"
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

// TestGraphiteStackerShellOut drives GraphiteStacker against a fake `gt` so the
// shell-out and output-parsing paths run without Graphite installed.
func TestGraphiteStackerShellOut(t *testing.T) {
	gt := testutil.FakeBin(t, "gt", `
case "$1" in
  sync) echo "stack synced" ;;
  log)  printf '%s\n' '◯ roa-2-top' '│' '◉ roa-1-base' '│' '◯ main' ;;
esac`)
	g := NewGraphiteStacker(gt)
	wt := Worktree{Path: t.TempDir()}

	branches, err := g.StackBranches(wt)
	require.NoError(t, err)
	require.Equal(t, []string{"roa-2-top", "roa-1-base"}, branches)

	st, err := g.Sync(wt)
	require.NoError(t, err)
	require.True(t, st.Rebased)
	require.Equal(t, "stack synced", st.Detail)
	require.Equal(t, branches, st.Branches)
}
