package command

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpawnCreatesPeer(t *testing.T) {
	ctx := testContext(t)
	var out strings.Builder
	require.NoError(t, doSpawn(ctx, "worker", "noop", "reviewer", ctx.RepoRoot, &out))
	require.Contains(t, out.String(), "Spawned peer")
}

func TestSpawnRequiresAgent(t *testing.T) {
	ctx := testContext(t)
	require.Error(t, doSpawn(ctx, "worker", "", "", ctx.RepoRoot, &strings.Builder{}),
		"spawn without --agent should fail")
}
