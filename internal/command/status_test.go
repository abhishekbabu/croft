package command

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusShowsWorktree(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "my-feature", "", "", &strings.Builder{}))

	var out strings.Builder
	require.NoError(t, doStatus(ctx, "my-feature", &out))
	require.Contains(t, out.String(), "my-feature")
}

func TestStatusUnknownWorktree(t *testing.T) {
	ctx := testContext(t)
	require.Error(t, doStatus(ctx, "nope", &strings.Builder{}),
		"status of an unknown worktree should fail")
}
