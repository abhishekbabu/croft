package provider

import (
	"testing"

	"github.com/abhishekbabu/croft/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestNoneInfraIsInert(t *testing.T) {
	wt := Worktree{Slug: "x", Path: "/tmp/x"}
	require.NoError(t, NoneInfra{}.Up(wt))
	require.NoError(t, NoneInfra{}.Down(wt))

	st, err := NoneInfra{}.Status(wt)
	require.NoError(t, err)
	require.False(t, st.Up, "NoneInfra.Status should report down")
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
