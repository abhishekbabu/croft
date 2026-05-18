package provider

import (
	"testing"

	"github.com/abhishekbabu/croft/internal/testutil"
	"github.com/stretchr/testify/require"
)

// TestPortlessRouterShellOut drives PortlessRouter against a fake `portless`.
func TestPortlessRouterShellOut(t *testing.T) {
	portless := testutil.FakeBin(t, "portless", `
case "$1" in
  get) echo "https://$2.localhost" ;;
esac
exit 0`)
	r := NewPortlessRouter(portless)
	wt := Worktree{Slug: "feat", Ports: map[string]int{"api": 3000}}

	url, err := r.Register(wt)
	require.NoError(t, err)
	require.Equal(t, "https://feat-api.localhost", url)

	require.NoError(t, r.Release(wt))
}
