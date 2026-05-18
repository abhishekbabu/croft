package command

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRmTearsDown(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "my-feature", "", "", "", &strings.Builder{}))
	wt, _, _ := ctx.Store.Get("my-feature")

	require.NoError(t, doRm(ctx, "my-feature", true, &strings.Builder{}))
	_, found, _ := ctx.Store.Get("my-feature")
	require.False(t, found, "registry should not contain the removed worktree")
	require.NoDirExists(t, wt.Path)
}

func TestRmIsIdempotent(t *testing.T) {
	ctx := testContext(t)
	var out strings.Builder
	require.NoError(t, doRm(ctx, "ghost", true, &out))
	require.Contains(t, out.String(), "Nothing to remove", "rm of an absent worktree should be a no-op")
}
