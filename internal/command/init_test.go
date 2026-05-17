package command

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
)

// newTestRepo creates an initialized git repo in a temp dir and returns it.
func newTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return dir
}

func TestInitYesCreatesValidConfig(t *testing.T) {
	dir := newTestRepo(t)
	var out strings.Builder
	if err := doInit(dir, false, true, strings.NewReader(""), &out); err != nil {
		t.Fatalf("doInit: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, config.ProjectFileName))
	if err != nil {
		t.Fatalf("read scaffolded config: %v", err)
	}
	if _, err := config.DecodeProject(data); err != nil {
		t.Fatalf("scaffolded config is invalid: %v", err)
	}
}

func TestInitRefusesExisting(t *testing.T) {
	dir := newTestRepo(t)
	var out strings.Builder
	if err := doInit(dir, false, true, strings.NewReader(""), &out); err != nil {
		t.Fatalf("first init: %v", err)
	}
	if err := doInit(dir, false, true, strings.NewReader(""), &out); err == nil {
		t.Fatal("second init without --force should fail")
	}
	if err := doInit(dir, true, true, strings.NewReader(""), &out); err != nil {
		t.Fatalf("init --force should overwrite: %v", err)
	}
}

func TestInitOutsideGitRepo(t *testing.T) {
	var out strings.Builder
	if err := doInit(t.TempDir(), false, true, strings.NewReader(""), &out); err == nil {
		t.Fatal("init outside a git repo should fail")
	}
}

func TestInitInteractiveAnswers(t *testing.T) {
	dir := newTestRepo(t)
	var out strings.Builder
	// name, worktree root, dev command, multiplexer, infra, router, stacker.
	answers := "myproj\n../trees\njust dev\ntmux\ndocker-compose\nnone\nnone\n"
	if err := doInit(dir, false, false, strings.NewReader(answers), &out); err != nil {
		t.Fatalf("doInit interactive: %v", err)
	}
	p, err := config.LoadProject(filepath.Join(dir, config.ProjectFileName))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if p.Project.Name != "myproj" {
		t.Errorf("name = %q, want myproj", p.Project.Name)
	}
	if p.Worktree.Naming != "myproj.{slug}" {
		t.Errorf("naming = %q, want myproj.{slug}", p.Worktree.Naming)
	}
	if p.Worktree.DevCommand != "just dev" {
		t.Errorf("dev_command = %q, want 'just dev'", p.Worktree.DevCommand)
	}
	if p.Providers.Multiplexer != "tmux" || p.Providers.Infra != "docker-compose" {
		t.Errorf("providers = %+v", p.Providers)
	}
}
