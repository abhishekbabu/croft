package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abhishekbabu/croft/internal/state"
	"github.com/stretchr/testify/require"
)

func TestNewCreatesWorktree(t *testing.T) {
	ctx := testContext(t)
	var out strings.Builder
	require.NoError(t, doNew(ctx, "my-feature", "", "", &out))

	wt, found, err := ctx.Store.Get("my-feature")
	require.NoError(t, err)
	require.True(t, found, "registry should contain the worktree")
	require.DirExists(t, wt.Path)
	require.Equal(t, 3000, wt.Ports["api"])
	require.Equal(t, 3001, wt.Ports["db"])
}

func TestNewIsIdempotent(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "my-feature", "", "", &strings.Builder{}))
	wt, _, _ := ctx.Store.Get("my-feature")

	// Re-running reconciles the existing worktree rather than failing.
	var out strings.Builder
	require.NoError(t, doNew(ctx, "my-feature", "", "", &out))
	require.Contains(t, out.String(), "Reconciling")
	again, _, _ := ctx.Store.Get("my-feature")
	require.Equal(t, wt.Ports, again.Ports, "reconcile must not change ports")
}

func TestNewAssignsDistinctPorts(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "first", "", "", &strings.Builder{}))
	require.NoError(t, doNew(ctx, "second", "", "", &strings.Builder{}))

	a, _, _ := ctx.Store.Get("first")
	b, _, _ := ctx.Store.Get("second")
	require.NotEqual(t, a.Ports["api"], b.Ports["api"], "second worktree should not reuse a port")
}

func TestNewWithAgent(t *testing.T) {
	ctx := testContext(t)
	var out strings.Builder
	// The "noop" exec agent runs `true`; with the none multiplexer it runs in
	// the foreground and exits cleanly.
	require.NoError(t, doNew(ctx, "agented", "", "noop", &out))
	require.Contains(t, out.String(), "Launched agent")

	rec, _, _ := ctx.Store.Get("agented")
	require.Equal(t, state.StatusWorking, rec.Status)
}

func TestNewWithUnknownAgent(t *testing.T) {
	ctx := testContext(t)
	require.Error(t, doNew(ctx, "x", "", "ghost", &strings.Builder{}),
		"doNew with an unknown agent should fail")
}

func TestNewStartsDevServer(t *testing.T) {
	ctx := testContext(t)
	var out strings.Builder
	require.NoError(t, doNew(ctx, "feat", "", "", &out))

	// With the none multiplexer croft can't host the dev server, so it prints
	// the command — with {port} substituted to the primary service's port.
	require.Contains(t, out.String(), "dev server")
	require.Contains(t, out.String(), "echo serving on 3000", "{port} should be substituted")
}

func TestSeedCopyFiles(t *testing.T) {
	repo := t.TempDir()
	wt := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".env.local"), []byte("SECRET=1"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "config"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "config", "local.toml"), []byte("x"), 0o600))

	var out strings.Builder
	seedCopyFiles(repo, wt, []string{".env.local", "config/local.toml", "missing.txt"}, &out)

	got, err := os.ReadFile(filepath.Join(wt, ".env.local"))
	require.NoError(t, err)
	require.Equal(t, "SECRET=1", string(got), "file content is copied")
	require.FileExists(t, filepath.Join(wt, "config", "local.toml"), "nested paths are created")
	require.NoFileExists(t, filepath.Join(wt, "missing.txt"), "a missing source is skipped silently")
	require.Contains(t, out.String(), "seeded: .env.local")
	require.NotContains(t, out.String(), "missing.txt")
}

func TestSeedCopyFilesReportsFailure(t *testing.T) {
	repo := t.TempDir()
	wt := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "sub"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "sub", "f"), []byte("x"), 0o600))
	// A plain file at wt/sub blocks creating the destination directory, so the
	// copy of an existing source genuinely fails.
	require.NoError(t, os.WriteFile(filepath.Join(wt, "sub"), nil, 0o600))

	var out strings.Builder
	seedCopyFiles(repo, wt, []string{"sub/f"}, &out)
	require.Contains(t, out.String(), "warning: could not seed sub/f")
}

func TestNewStartsDevServerInSession(t *testing.T) {
	ctx := testContext(t)
	// A managed multiplexer hosts the dev server in a window.
	ctx.Providers.Multiplexer = fakeMux{managed: true}

	var out strings.Builder
	require.NoError(t, doNew(ctx, "feat", "", "", &out))
	require.Contains(t, out.String(), "dev server: started")
}
