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

// fakeMux is a controllable Multiplexer for status-derivation tests.
type fakeMux struct {
	managed bool
	windows map[string]bool // "session:window" -> present
}

func (f fakeMux) Managed() bool                                         { return f.managed }
func (f fakeMux) CreateSession(string, string, map[string]string) error { return nil }
func (f fakeMux) RunWindow(string, string, string, map[string]string, []string) error {
	return nil
}
func (f fakeMux) HasWindow(name, window string) bool              { return f.windows[name+":"+window] }
func (f fakeMux) Attach(string) error                             { return nil }
func (f fakeMux) Kill(string) error                               { return nil }
func (f fakeMux) CapturePane(string, string, int) (string, error) { return "", nil }

// setupRepo builds a git repo (one commit) with a croft.toml and isolated XDG
// dirs, and returns the repo path.
func setupRepo(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	require.NoError(t, os.MkdirAll(repo, 0o755))
	testutil.InitGitRepo(t, repo)

	cfg := `
[project]
name = "demo"
[worktree]
root = "../wt"
dev_command = "echo serving on {port}"
[ports]
range = "3000-3999"
services = ["api", "db"]
[[agents]]
name = "noop"
runner = "exec"
command = ["true"]
`
	require.NoError(t, os.WriteFile(filepath.Join(repo, "croft.toml"), []byte(cfg), 0o644))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "cfg"))
	return repo
}

// testContext sets up a repo and returns its loaded context.
func testContext(t *testing.T) *appContext {
	t.Helper()
	ctx, err := loadContext(setupRepo(t))
	require.NoError(t, err, "loadContext")
	return ctx
}

func TestNewLsStatusRm(t *testing.T) {
	ctx := testContext(t)

	var out strings.Builder
	require.NoError(t, doNew(ctx, "my-feature", "", "", &out))

	wt, found, err := ctx.Store.Get("my-feature")
	require.NoError(t, err)
	require.True(t, found, "registry should contain the worktree")
	require.DirExists(t, wt.Path)
	require.Equal(t, 3000, wt.Ports["api"])
	require.Equal(t, 3001, wt.Ports["db"])

	// new is idempotent — re-running reconciles the existing worktree.
	out.Reset()
	require.NoError(t, doNew(ctx, "my-feature", "", "", &out))
	require.Contains(t, out.String(), "Reconciling")
	again, _, _ := ctx.Store.Get("my-feature")
	require.Equal(t, wt.Ports, again.Ports, "reconcile must not change ports")

	// A second worktree gets a distinct port set.
	require.NoError(t, doNew(ctx, "other", "", "", &strings.Builder{}))
	other, _, _ := ctx.Store.Get("other")
	require.NotEqual(t, wt.Ports["api"], other.Ports["api"], "second worktree should not reuse a port")

	out.Reset()
	require.NoError(t, doLs(ctx, &out))
	require.Contains(t, out.String(), "my-feature")
	require.Contains(t, out.String(), "other")

	out.Reset()
	require.NoError(t, doStatus(ctx, "my-feature", &out))
	require.Contains(t, out.String(), "my-feature")

	// rm tears down and is idempotent.
	require.NoError(t, doRm(ctx, "my-feature", true, &strings.Builder{}))
	_, found, _ = ctx.Store.Get("my-feature")
	require.False(t, found, "registry should not contain the removed worktree")
	require.NoDirExists(t, wt.Path)

	out.Reset()
	require.NoError(t, doRm(ctx, "my-feature", true, &out))
	require.Contains(t, out.String(), "Nothing to remove", "repeat rm should be a no-op")
}

func TestStatusUnknownWorktree(t *testing.T) {
	ctx := testContext(t)
	require.Error(t, doStatus(ctx, "nope", &strings.Builder{}), "status of an unknown worktree should fail")
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
	require.NoError(t, doNew(ctx, "feat", "", "", &strings.Builder{}))
	rec, _, _ := ctx.Store.Get("feat")

	require.Equal(t, "-", ctx.liveStatus(rec), "no agent")

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

func TestSpawnAndFleet(t *testing.T) {
	ctx := testContext(t)

	var out strings.Builder
	require.NoError(t, doSpawn(ctx, "worker", "noop", "reviewer", ctx.RepoRoot, &out))
	require.Contains(t, out.String(), "Spawned peer")

	out.Reset()
	require.NoError(t, doFleetStatus(ctx, &out))
	require.Contains(t, out.String(), "worker", "fleet status should list the peer")

	out.Reset()
	require.NoError(t, doFleetMsg(ctx, "worker", "ship it", &out))
	require.Contains(t, out.String(), "Delivered")

	require.Error(t, doFleetMsg(ctx, "ghost", "hello", &strings.Builder{}),
		"dispatch to an unknown peer should fail")
}

func TestSpawnRequiresAgent(t *testing.T) {
	ctx := testContext(t)
	require.Error(t, doSpawn(ctx, "worker", "", "", ctx.RepoRoot, &strings.Builder{}),
		"spawn without --agent should fail")
}

func TestLoadContextWithoutConfig(t *testing.T) {
	dir := testutil.EmptyGitRepo(t)
	_, err := loadContext(dir)
	require.Error(t, err, "loadContext without croft.toml should fail")
}
