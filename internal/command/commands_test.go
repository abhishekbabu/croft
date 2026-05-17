package command

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupRepo builds a git repo (one commit) with a croft.toml and isolated XDG
// dirs, and returns the repo path.
func setupRepo(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init")
	if err := os.WriteFile(filepath.Join(repo, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-m", "init")

	cfg := `
[project]
name = "demo"
[worktree]
root = "../wt"
[ports]
range = "3000-3999"
services = ["api", "db"]
[[agents]]
name = "noop"
runner = "exec"
command = ["true"]
`
	if err := os.WriteFile(filepath.Join(repo, "croft.toml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "cfg"))
	return repo
}

func TestNewLsStatusRm(t *testing.T) {
	repo := setupRepo(t)
	ctx, err := loadContext(repo)
	if err != nil {
		t.Fatalf("loadContext: %v", err)
	}

	var out strings.Builder
	if err := doNew(ctx, "my-feature", "", "", &out); err != nil {
		t.Fatalf("doNew: %v", err)
	}
	wt, found, err := ctx.Store.Get("my-feature")
	if err != nil || !found {
		t.Fatalf("registry missing worktree: found=%v err=%v", found, err)
	}
	if !dirExists(wt.Path) {
		t.Fatalf("worktree directory not created: %s", wt.Path)
	}
	if wt.Ports["api"] != 3000 || wt.Ports["db"] != 3001 {
		t.Errorf("ports = %v, want api=3000 db=3001", wt.Ports)
	}

	// new is idempotent — re-running reconciles the existing worktree.
	out.Reset()
	if err := doNew(ctx, "my-feature", "", "", &out); err != nil {
		t.Fatalf("doNew (repeat): %v", err)
	}
	if !strings.Contains(out.String(), "Reconciling") {
		t.Errorf("repeat doNew should reconcile, got %q", out.String())
	}
	if again, _, _ := ctx.Store.Get("my-feature"); again.Ports["api"] != wt.Ports["api"] {
		t.Errorf("reconcile must not change ports: %v vs %v", again.Ports, wt.Ports)
	}

	// A second worktree gets a distinct port set.
	if err := doNew(ctx, "other", "", "", &strings.Builder{}); err != nil {
		t.Fatalf("doNew other: %v", err)
	}
	other, _, _ := ctx.Store.Get("other")
	if other.Ports["api"] == wt.Ports["api"] {
		t.Errorf("second worktree reused port %d", other.Ports["api"])
	}

	out.Reset()
	if err := doLs(ctx, &out); err != nil {
		t.Fatalf("doLs: %v", err)
	}
	if !strings.Contains(out.String(), "my-feature") || !strings.Contains(out.String(), "other") {
		t.Errorf("doLs output missing worktrees:\n%s", out.String())
	}

	out.Reset()
	if err := doStatus(ctx, "my-feature", &out); err != nil {
		t.Fatalf("doStatus: %v", err)
	}
	if !strings.Contains(out.String(), "my-feature") {
		t.Errorf("doStatus output:\n%s", out.String())
	}

	// rm tears down and is idempotent.
	if err := doRm(ctx, "my-feature", true, &strings.Builder{}); err != nil {
		t.Fatalf("doRm: %v", err)
	}
	if _, found, _ := ctx.Store.Get("my-feature"); found {
		t.Error("registry should not contain removed worktree")
	}
	if dirExists(wt.Path) {
		t.Error("worktree directory should be gone after rm")
	}
	out.Reset()
	if err := doRm(ctx, "my-feature", true, &out); err != nil {
		t.Fatalf("doRm (repeat): %v", err)
	}
	if !strings.Contains(out.String(), "Nothing to remove") {
		t.Errorf("repeat doRm should be a no-op, got %q", out.String())
	}
}

func TestStatusUnknownWorktree(t *testing.T) {
	repo := setupRepo(t)
	ctx, err := loadContext(repo)
	if err != nil {
		t.Fatalf("loadContext: %v", err)
	}
	if err := doStatus(ctx, "nope", &strings.Builder{}); err == nil {
		t.Error("status of unknown worktree should error")
	}
}

func TestNewWithAgent(t *testing.T) {
	repo := setupRepo(t)
	ctx, err := loadContext(repo)
	if err != nil {
		t.Fatalf("loadContext: %v", err)
	}
	var out strings.Builder
	// The "noop" exec agent runs `true`; with the none multiplexer it runs in
	// the foreground and exits cleanly.
	if err := doNew(ctx, "agented", "", "noop", &out); err != nil {
		t.Fatalf("doNew --agent: %v", err)
	}
	if !strings.Contains(out.String(), "Launched agent") {
		t.Errorf("expected agent launch in output:\n%s", out.String())
	}
	rec, _, _ := ctx.Store.Get("agented")
	if rec.Status != "working" {
		t.Errorf("status = %q, want working", rec.Status)
	}
}

func TestNewWithUnknownAgent(t *testing.T) {
	repo := setupRepo(t)
	ctx, err := loadContext(repo)
	if err != nil {
		t.Fatalf("loadContext: %v", err)
	}
	if err := doNew(ctx, "x", "", "ghost", &strings.Builder{}); err == nil {
		t.Error("doNew with an unknown agent should error")
	}
}

func TestSyncNoStacker(t *testing.T) {
	repo := setupRepo(t)
	ctx, err := loadContext(repo)
	if err != nil {
		t.Fatalf("loadContext: %v", err)
	}
	if err := doNew(ctx, "feat", "", "", &strings.Builder{}); err != nil {
		t.Fatalf("doNew: %v", err)
	}
	var out strings.Builder
	if err := doSync(ctx, "", false, &out); err != nil {
		t.Fatalf("doSync: %v", err)
	}
	if !strings.Contains(out.String(), "feat") || !strings.Contains(out.String(), "synced") {
		t.Errorf("sync output:\n%s", out.String())
	}
}

func TestSyncRefusesMidRebase(t *testing.T) {
	repo := setupRepo(t)
	ctx, err := loadContext(repo)
	if err != nil {
		t.Fatalf("loadContext: %v", err)
	}
	if err := doNew(ctx, "feat", "", "", &strings.Builder{}); err != nil {
		t.Fatalf("doNew: %v", err)
	}
	rec, _, _ := ctx.Store.Get("feat")
	gd, err := ctx.Manager.GitDir(rec.Path)
	if err != nil {
		t.Fatalf("GitDir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(gd, "rebase-merge"), 0o755); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := doSync(ctx, "feat", false, &out); err == nil {
		t.Error("doSync should fail when a worktree is mid-rebase")
	}
	if !strings.Contains(out.String(), "mid-rebase") {
		t.Errorf("sync output should mention mid-rebase:\n%s", out.String())
	}
}

func TestLoadContextWithoutConfig(t *testing.T) {
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if _, err := loadContext(dir); err == nil {
		t.Error("loadContext without croft.toml should error")
	}
}
