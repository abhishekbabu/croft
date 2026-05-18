package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
	"github.com/abhishekbabu/croft/internal/provider"
	"github.com/abhishekbabu/croft/internal/state"
	"github.com/abhishekbabu/croft/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestLoadContextWithoutConfig(t *testing.T) {
	dir := testutil.EmptyGitRepo(t)
	_, err := loadContext(dir)
	require.Error(t, err, "loadContext without croft.toml should fail")
}

func TestDevCommandSubstitution(t *testing.T) {
	ctx := &appContext{Config: config.ProjectConfig{
		Worktree: config.WorktreeSection{DevCommand: "run --port {port}"},
		Ports:    config.PortsSection{Services: []string{"api", "db"}},
	}}
	rec := state.Worktree{Ports: map[string]int{"api": 5000, "db": 5001}}
	require.Equal(t, "run --port 5000", ctx.devCommand(rec))
	require.Empty(t, (&appContext{}).devCommand(rec), "no dev_command yields empty")
}

func TestLiveStatusDerivation(t *testing.T) {
	ctx := testContext(t)
	require.NoError(t, doNew(ctx, "feat", "", "", "", &strings.Builder{}))
	rec, _, _ := ctx.Store.Get("feat")

	require.Empty(t, ctx.liveStatus(rec), "no agent yields the empty status")

	session := provider.ProjectName(ctx.providerWorktree(rec))
	rec.Status = state.StatusWorking

	// Agent window present -> working.
	ctx.Providers.Multiplexer = fakeMux{managed: true, windows: map[string]bool{session + ":agent": true}}
	require.Equal(t, state.StatusWorking, ctx.liveStatus(rec), "agent window present")

	// Agent window gone -> done.
	ctx.Providers.Multiplexer = fakeMux{managed: true, windows: map[string]bool{}}
	require.Equal(t, state.StatusDone, ctx.liveStatus(rec), "agent window gone")

	// A rebase in progress overrides everything.
	gd, err := ctx.Manager.GitDir(rec.Path)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(gd, "rebase-merge"), 0o755))
	require.Equal(t, state.StatusRebase, ctx.liveStatus(rec), "mid-rebase overrides")
}
