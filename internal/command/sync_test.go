package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyncNoStacker(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "feat", "", "", &strings.Builder{}))

	var out strings.Builder
	require.NoError(t, doSync(ctx, "", false, &out))
	require.Contains(t, out.String(), "feat")
	require.Contains(t, out.String(), "synced")
}

func TestSyncRefusesMidRebase(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "feat", "", "", &strings.Builder{}))

	rec, _, _ := ctx.Store.Get("feat")
	gd, err := ctx.Manager.GitDir(rec.Path)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(gd, "rebase-merge"), 0o755))

	var out strings.Builder
	require.Error(t, doSync(ctx, "feat", false, &out), "sync should fail mid-rebase")
	require.Contains(t, out.String(), "mid-rebase")
}
