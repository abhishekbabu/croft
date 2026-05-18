// Package testutil provides shared fixtures for croft's tests. Several
// packages need an identically set-up git repository; defining it once here
// keeps the setup consistent and the test files focused on what they assert.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// gitIdentity is a deterministic author/committer identity, so test commits
// never depend on (or mutate) the developer's global git config.
var gitIdentity = []string{
	"GIT_AUTHOR_NAME=croft-test", "GIT_AUTHOR_EMAIL=test@croft",
	"GIT_COMMITTER_NAME=croft-test", "GIT_COMMITTER_EMAIL=test@croft",
}

// git runs a git subcommand in dir, failing the test on any error.
func git(t testing.TB, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), gitIdentity...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v\n%s", args, out)
}

// EmptyGitRepo runs `git init` in a fresh temp dir and returns its path. The
// repo has no commits.
func EmptyGitRepo(t testing.TB) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init")
	return dir
}

// GitRepo creates a git repo with a single commit on the default branch in a
// fresh temp dir and returns its path.
func GitRepo(t testing.TB) string {
	t.Helper()
	dir := t.TempDir()
	InitGitRepo(t, dir)
	return dir
}

// InitGitRepo turns an existing directory into a git repo with one commit.
// Use it when the repo must live at a specific path (e.g. a subdirectory whose
// siblings the test also relies on).
func InitGitRepo(t testing.TB, dir string) {
	t.Helper()
	git(t, dir, "init")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644))
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", "init")
}
