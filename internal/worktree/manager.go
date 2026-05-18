package worktree

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/abhishekbabu/croft/internal/sh"
)

// GitWorktree is one entry parsed from `git worktree list --porcelain`.
type GitWorktree struct {
	Path     string
	Branch   string // empty when detached
	Detached bool
	Bare     bool
}

// Manager wraps the `git worktree` plumbing for one repository. It never
// reimplements git — every operation shells out (PLAN.md §11).
type Manager struct {
	repoRoot string
}

// NewManager returns a Manager rooted at the given repository root.
func NewManager(repoRoot string) *Manager {
	return &Manager{repoRoot: repoRoot}
}

// git runs a git subcommand at the repository root and returns stdout.
func (m *Manager) git(args ...string) (string, error) {
	return gitAt(m.repoRoot, args...)
}

// gitAt runs a git subcommand in dir and returns stdout.
func gitAt(dir string, args ...string) (string, error) {
	return sh.Capture("git", dir, nil, args...)
}

// GitDir returns the absolute .git directory for the worktree at path.
func (m *Manager) GitDir(path string) (string, error) {
	out, err := gitAt(path, "rev-parse", "--git-dir")
	if err != nil {
		return "", err
	}
	gd := strings.TrimSpace(out)
	if !filepath.IsAbs(gd) {
		gd = filepath.Join(path, gd)
	}
	return gd, nil
}

// InRebase reports whether the worktree at path has a rebase in progress —
// either a `git rebase` or a Graphite restack that hit conflicts.
func (m *Manager) InRebase(path string) bool {
	gd, err := m.GitDir(path)
	if err != nil {
		return false
	}
	for _, d := range []string{"rebase-merge", "rebase-apply"} {
		if info, err := os.Stat(filepath.Join(gd, d)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

// isDirty reports whether the worktree at path has uncommitted changes in the
// working tree or the index.
func (m *Manager) isDirty(path string) bool {
	_, worktreeErr := sh.Capture("git", path, nil, "diff", "--quiet")
	_, indexErr := sh.Capture("git", path, nil, "diff", "--cached", "--quiet")
	return worktreeErr != nil || indexErr != nil
}

// Stash saves uncommitted changes (including untracked files) and reports
// whether anything was stashed.
func (m *Manager) Stash(path, msg string) (bool, error) {
	if !m.isDirty(path) {
		return false, nil
	}
	if _, err := gitAt(path, "stash", "push", "-m", msg, "--include-untracked"); err != nil {
		return false, err
	}
	return true, nil
}

// StashPop restores the most recently stashed changes.
func (m *Manager) StashPop(path string) error {
	_, err := gitAt(path, "stash", "pop")
	return err
}

// branchExists reports whether branch is a known local branch.
func (m *Manager) branchExists(branch string) bool {
	_, err := m.git("rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	return err == nil
}

// Add creates a worktree at path. An existing branch is checked out; otherwise
// a new branch is created from startPoint (HEAD when startPoint is empty).
func (m *Manager) Add(path, branch, startPoint string) error {
	args := []string{"worktree", "add"}
	if m.branchExists(branch) {
		args = append(args, path, branch)
	} else {
		args = append(args, "-b", branch, path)
		if startPoint != "" {
			args = append(args, startPoint)
		}
	}
	_, err := m.git(args...)
	return err
}

// Remove removes the worktree at path. force removes it even when dirty.
func (m *Manager) Remove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	_, err := m.git(append(args, path)...)
	return err
}

// Prune runs `git worktree prune`, clearing administrative entries for
// worktrees whose directories no longer exist.
func (m *Manager) Prune() error {
	_, err := m.git("worktree", "prune")
	return err
}

// List parses `git worktree list --porcelain`.
func (m *Manager) List() ([]GitWorktree, error) {
	out, err := m.git("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreeList(out), nil
}

// parseWorktreeList parses porcelain `git worktree list` output. Records are
// separated by blank lines.
func parseWorktreeList(out string) []GitWorktree {
	var list []GitWorktree
	var cur *GitWorktree
	flush := func() {
		if cur != nil {
			list = append(list, *cur)
			cur = nil
		}
	}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			flush()
			continue
		}
		key, val, _ := strings.Cut(line, " ")
		if key == "worktree" {
			flush()
			cur = &GitWorktree{Path: val}
			continue
		}
		if cur == nil {
			continue
		}
		switch key {
		case "branch":
			cur.Branch = strings.TrimPrefix(val, "refs/heads/")
		case "detached":
			cur.Detached = true
		case "bare":
			cur.Bare = true
		}
	}
	flush()
	return list
}
