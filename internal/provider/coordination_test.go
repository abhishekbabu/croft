package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoopCoordinationIsInert(t *testing.T) {
	var c NoopCoordination

	_, err := c.Spawn(PeerSpec{Name: "x"})
	require.Error(t, err, "the noop coordination provider rejects Spawn")

	peers, err := c.Status()
	require.NoError(t, err)
	require.Empty(t, peers, "the noop coordination provider has no peers")

	require.Error(t, c.Dispatch(Peer{Name: "x"}, "msg"),
		"the noop coordination provider rejects Dispatch")
}
