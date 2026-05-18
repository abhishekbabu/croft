package command

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLsEmpty(t *testing.T) {
	ctx := testContext(t)
	var out strings.Builder
	require.NoError(t, doLs(ctx, &out))
	require.Contains(t, out.String(), "No worktrees yet")
}

func TestLsListsWorktrees(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "my-feature", "", "", "", &strings.Builder{}))
	require.NoError(t, doNew(ctx, "other", "", "", "", &strings.Builder{}))

	var out strings.Builder
	require.NoError(t, doLs(ctx, &out))
	require.Contains(t, out.String(), "my-feature")
	require.Contains(t, out.String(), "other")
}
