package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoneRouterIsInert(t *testing.T) {
	wt := Worktree{Slug: "x", Path: "/tmp/x"}
	url, err := NoneRouter{}.Register(wt)
	require.NoError(t, err)
	require.Empty(t, url, "NoneRouter.Register should yield no URL")
	require.NoError(t, NoneRouter{}.Release(wt))
}
