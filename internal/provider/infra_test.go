package provider

import (
	"testing"

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
