package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abhishekbabu/croft/internal/provider"
	"github.com/stretchr/testify/require"
)

// resolvedStacker is a Stacker whose stack is always fully resolved, so the
// teardown gate in syncOne fires.
type resolvedStacker struct{}

func (resolvedStacker) Sync(provider.Worktree) (provider.StackState, error) {
	return provider.StackState{Rebased: true, Detail: "synced"}, nil
}
func (resolvedStacker) StackBranches(provider.Worktree) ([]string, error) {
	return []string{"feat"}, nil
}
func (resolvedStacker) AllResolved(provider.Worktree) (bool, error) { return true, nil }

func TestSyncNoStacker(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "feat", "", "", "", &strings.Builder{}))

	var out strings.Builder
	require.NoError(t, doSync(ctx, "", false, &out))
	require.Contains(t, out.String(), "feat")
	require.Contains(t, out.String(), "synced")
}

func TestSyncPrunesResolvedStack(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "feat", "", "", "", &strings.Builder{}))
	rec, _, _ := ctx.Store.Get("feat")

	// A fully-resolved stack plus --prune trips the teardown gate.
	ctx.Providers.Stacker = resolvedStacker{}

	var out strings.Builder
	require.NoError(t, doSync(ctx, "feat", true, &out))
	require.Contains(t, out.String(), "fully resolved")

	_, found, _ := ctx.Store.Get("feat")
	require.False(t, found, "a resolved stack should be torn down under --prune")
	require.NoDirExists(t, rec.Path)
}

func TestSyncRefusesMidRebase(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "feat", "", "", "", &strings.Builder{}))

	rec, _, _ := ctx.Store.Get("feat")
	gd, err := ctx.Manager.GitDir(rec.Path)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(gd, "rebase-merge"), 0o755))

	var out strings.Builder
	require.Error(t, doSync(ctx, "feat", false, &out), "sync should fail mid-rebase")
	require.Contains(t, out.String(), "mid-rebase")
}
