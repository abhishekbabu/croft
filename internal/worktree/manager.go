package worktree

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// GitWorktree is one entry parsed from `git worktree list --porcelain`.
type GitWorktree struct {
	Path     string
	Branch   string // empty when detached
	Head     string
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

// git runs a git subcommand inside the repository and returns stdout.
func (m *Manager) git(args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", m.repoRoot}, args...)...)
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return out.String(), nil
}

// BranchExists reports whether branch is a known local branch.
func (m *Manager) BranchExists(branch string) bool {
	_, err := m.git("rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	return err == nil
}

// Add creates a worktree at path. An existing branch is checked out; otherwise
// a new branch is created from startPoint (HEAD when startPoint is empty).
func (m *Manager) Add(path, branch, startPoint string) error {
	args := []string{"worktree", "add"}
	if m.BranchExists(branch) {
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
		case "HEAD":
			cur.Head = val
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
