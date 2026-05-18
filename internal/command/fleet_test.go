package command

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFleetStatusListsPeers(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doSpawn(ctx, "worker", "noop", "reviewer", ctx.RepoRoot, &strings.Builder{}))

	var out strings.Builder
	require.NoError(t, doFleetStatus(ctx, &out))
	require.Contains(t, out.String(), "worker", "fleet status should list the peer")
}

func TestFleetMsgDelivers(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doSpawn(ctx, "worker", "noop", "reviewer", ctx.RepoRoot, &strings.Builder{}))

	var out strings.Builder
	require.NoError(t, doFleetMsg(ctx, "worker", "ship it", &out))
	require.Contains(t, out.String(), "Delivered")

	require.Error(t, doFleetMsg(ctx, "ghost", "hello", &strings.Builder{}),
		"dispatch to an unknown peer should fail")
}
