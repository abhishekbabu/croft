package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoneStackerIsInert(t *testing.T) {
	wt := Worktree{Slug: "x", Path: "/tmp/x"}

	ok, err := NoneStacker{}.AllResolved(wt)
	require.NoError(t, err)
	require.False(t, ok, "NoneStacker.AllResolved must be false")

	st, err := NoneStacker{}.Sync(wt)
	require.NoError(t, err)
	require.False(t, st.Rebased, "NoneStacker.Sync rebases nothing")

	branches, err := NoneStacker{}.StackBranches(wt)
	require.NoError(t, err)
	require.Empty(t, branches)
}
