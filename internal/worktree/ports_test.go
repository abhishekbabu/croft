package worktree

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllocatePortsLowestFree(t *testing.T) {
	got, err := AllocatePorts(3000, 3999, []string{"api", "db"}, nil)
	require.NoError(t, err)
	require.Equal(t, map[string]int{"api": 3000, "db": 3001}, got)
}

func TestAllocatePortsSkipsTaken(t *testing.T) {
	taken := map[int]bool{3000: true, 3001: true}
	got, err := AllocatePorts(3000, 3999, []string{"api"}, taken)
	require.NoError(t, err)
	require.Equal(t, 3002, got["api"], "should skip taken ports")
	require.NotContains(t, taken, 3002, "must not mutate the taken set")
}

func TestAllocatePortsExhausted(t *testing.T) {
	_, err := AllocatePorts(3000, 3001, []string{"a", "b", "c"}, nil)
	require.Error(t, err, "more services than ports should exhaust the range")
}
