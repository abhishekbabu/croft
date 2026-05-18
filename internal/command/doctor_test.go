package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoctorClean(t *testing.T) {
	ctx := testContext(t)
	var out strings.Builder
	require.NoError(t, doDoctor(ctx, false, &out))
	require.Contains(t, out.String(), "All clear")
}

func TestDoctorStaleRegistry(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "feat", "", "", &strings.Builder{}))

	rec, _, _ := ctx.Store.Get("feat")
	// Simulate the worktree directory vanishing out from under croft.
	require.NoError(t, os.RemoveAll(rec.Path))

	var out strings.Builder
	require.NoError(t, doDoctor(ctx, false, &out))
	require.Contains(t, out.String(), "issue", "doctor should detect the stale entry")

	require.NoError(t, doDoctor(ctx, true, &strings.Builder{}))
	_, found, _ := ctx.Store.Get("feat")
	require.False(t, found, "stale registry entry should be gone after doctor --fix")
}

func TestDoctorOrphanDir(t *testing.T) {
	ctx := testContext(t)
	// A directory matching the naming pattern but with no registry entry.
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.WorktreeRoot, "demo.ghost"), 0o755))

	var out strings.Builder
	require.NoError(t, doDoctor(ctx, false, &out))
	require.Contains(t, out.String(), "orphan", "doctor should detect the orphan directory")
}
