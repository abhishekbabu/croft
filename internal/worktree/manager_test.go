package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseWorktreeList(t *testing.T) {
	out := "worktree /repo\n" +
		"HEAD abc123\n" +
		"branch refs/heads/main\n" +
		"\n" +
		"worktree /repo/../wt/demo.feat\n" +
		"HEAD def456\n" +
		"detached\n"
	got := parseWorktreeList(out)
	if len(got) != 2 {
		t.Fatalf("parsed %d worktrees, want 2", len(got))
	}
	if got[0].Branch != "main" || got[0].Path != "/repo" {
		t.Errorf("first entry = %+v", got[0])
	}
	if !got[1].Detached || got[1].Branch != "" {
		t.Errorf("second entry should be detached: %+v", got[1])
	}
}

// initRepo creates a git repo with one commit on the default branch.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-m", "init")
	return dir
}

func TestManagerAddListRemove(t *testing.T) {
	repo := initRepo(t)
	mgr := NewManager(repo)
	wtPath := filepath.Join(t.TempDir(), "demo.feat")

	if mgr.BranchExists("feat") {
		t.Fatal("feat branch should not exist yet")
	}
	if err := mgr.Add(wtPath, "feat", ""); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !mgr.BranchExists("feat") {
		t.Error("feat branch should exist after Add")
	}
	if _, err := os.Stat(wtPath); err != nil {
		t.Errorf("worktree directory missing: %v", err)
	}

	list, err := mgr.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var found bool
	for _, w := range list {
		if w.Branch == "feat" {
			found = true
		}
	}
	if !found {
		t.Errorf("List did not include the feat worktree: %+v", list)
	}

	if err := mgr.Remove(wtPath, false); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Errorf("worktree directory should be gone after Remove")
	}
	if err := mgr.Prune(); err != nil {
		t.Errorf("Prune: %v", err)
	}
}
