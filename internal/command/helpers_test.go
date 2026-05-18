package command

import (
	"os"
	"path/filepath"
	"testing"

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
