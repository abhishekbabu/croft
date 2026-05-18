package provider

import (
	"testing"

	"github.com/abhishekbabu/croft/internal/testutil"
	"github.com/stretchr/testify/require"
)

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

// TestComposeInfraShellOut drives ComposeInfra against a fake `docker`.
func TestComposeInfraShellOut(t *testing.T) {
	wt := Worktree{Slug: "feat", Path: t.TempDir()}

	running := testutil.FakeBin(t, "docker", `
for a in "$@"; do
  [ "$a" = "ps" ] && { echo abc123; echo def456; }
done
exit 0`)
	c := NewComposeInfra(running)
	require.NoError(t, c.Up(wt))
	require.NoError(t, c.Down(wt))

	st, err := c.Status(wt)
	require.NoError(t, err)
	require.True(t, st.Up, "two container ids should report the stack up")
	require.Equal(t, "2 containers", st.Detail)

	empty := testutil.FakeBin(t, "docker", "exit 0")
	st, err = NewComposeInfra(empty).Status(wt)
	require.NoError(t, err)
	require.False(t, st.Up, "no container ids should report the stack down")
}

// TestPortlessRouterShellOut drives PortlessRouter against a fake `portless`.
func TestPortlessRouterShellOut(t *testing.T) {
	portless := testutil.FakeBin(t, "portless", `
case "$1" in
  get) echo "https://$2.localhost" ;;
esac
exit 0`)
	r := NewPortlessRouter(portless)
	wt := Worktree{Slug: "feat", Ports: map[string]int{"api": 3000}}

	url, err := r.Register(wt)
	require.NoError(t, err)
	require.Equal(t, "https://feat-api.localhost", url)

	require.NoError(t, r.Release(wt))
}
