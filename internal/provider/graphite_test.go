package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/abhishekbabu/croft/internal/testutil"
	"github.com/stretchr/testify/require"
)

// prependToPATH puts binPath's directory at the front of $PATH for the test,
// so a hardcoded tool lookup (loadPRStates resolves "gh") finds the fake.
func prependToPATH(t *testing.T, binPath string) {
	t.Helper()
	t.Setenv("PATH", filepath.Dir(binPath)+string(os.PathListSeparator)+os.Getenv("PATH"))
}

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

// TestGraphiteAllResolved drives the teardown-gate logic against a fake `gt`
// (the stack) and a fake `gh` (the PR states loadPRStates queries).
func TestGraphiteAllResolved(t *testing.T) {
	const gtLog = `[ "$1" = log ] && printf '%s\n' 'roa-1' 'roa-2'
exit 0`
	wt := Worktree{Path: t.TempDir()}

	t.Run("all merged or closed is resolved", func(t *testing.T) {
		gt := testutil.FakeBin(t, "gt", gtLog)
		gh := testutil.FakeBin(t, "gh",
			`echo '[{"headRefName":"roa-1","state":"MERGED"},{"headRefName":"roa-2","state":"CLOSED"}]'`)
		prependToPATH(t, gh)

		ok, err := NewGraphiteStacker(gt).AllResolved(wt)
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("an open branch is not resolved", func(t *testing.T) {
		gt := testutil.FakeBin(t, "gt", gtLog)
		gh := testutil.FakeBin(t, "gh",
			`echo '[{"headRefName":"roa-1","state":"MERGED"},{"headRefName":"roa-2","state":"OPEN"}]'`)
		prependToPATH(t, gh)

		ok, err := NewGraphiteStacker(gt).AllResolved(wt)
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("an empty stack is not resolved", func(t *testing.T) {
		gt := testutil.FakeBin(t, "gt", `[ "$1" = log ] && echo 'main'
exit 0`)
		ok, err := NewGraphiteStacker(gt).AllResolved(wt)
		require.NoError(t, err)
		require.False(t, ok, "an undeterminable stack must never trip the gate")
	})
}
