package command

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
	"github.com/stretchr/testify/require"
)

// newTestRepo creates an initialized git repo in a temp dir and returns it.
func newTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	out, err := exec.Command("git", "-C", dir, "init").CombinedOutput()
	require.NoError(t, err, "git init\n%s", out)
	return dir
}

func TestInitYesCreatesValidConfig(t *testing.T) {
	dir := newTestRepo(t)
	var out strings.Builder
	require.NoError(t, doInit(dir, false, true, strings.NewReader(""), &out))

	data, err := os.ReadFile(filepath.Join(dir, config.ProjectFileName))
	require.NoError(t, err)
	_, err = config.DecodeProject(data)
	require.NoError(t, err, "scaffolded config should be valid")
}

func TestInitRefusesExisting(t *testing.T) {
	dir := newTestRepo(t)
	var out strings.Builder
	require.NoError(t, doInit(dir, false, true, strings.NewReader(""), &out))
	require.Error(t, doInit(dir, false, true, strings.NewReader(""), &out),
		"second init without --force should fail")
	require.NoError(t, doInit(dir, true, true, strings.NewReader(""), &out),
		"init --force should overwrite")
}

func TestInitOutsideGitRepo(t *testing.T) {
	var out strings.Builder
	require.Error(t, doInit(t.TempDir(), false, true, strings.NewReader(""), &out),
		"init outside a git repo should fail")
}

func TestInitInteractiveAnswers(t *testing.T) {
	dir := newTestRepo(t)
	var out strings.Builder
	// name, worktree root, dev command, multiplexer, infra, router, stacker.
	answers := "myproj\n../trees\njust dev\ntmux\ndocker-compose\nnone\nnone\n"
	require.NoError(t, doInit(dir, false, false, strings.NewReader(answers), &out))

	p, err := config.LoadProject(filepath.Join(dir, config.ProjectFileName))
	require.NoError(t, err)
	require.Equal(t, "myproj", p.Project.Name)
	require.Equal(t, "myproj.{slug}", p.Worktree.Naming)
	require.Equal(t, "just dev", p.Worktree.DevCommand)
	require.Equal(t, config.MultiplexerTmux, p.Providers.Multiplexer)
	require.Equal(t, config.InfraDockerCompose, p.Providers.Infra)
}
