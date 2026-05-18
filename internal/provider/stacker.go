package provider

import (
	"regexp"
	"strings"

	"github.com/abhishekbabu/croft/internal/sh"
)

// Stacker syncs a worktree's branch stack against the trunk and reports stack
// resolution (used by `croft sync` and the teardown gate).
type Stacker interface {
	// Sync rebases the worktree's stack onto the trunk.
	Sync(wt Worktree) (StackState, error)
	// StackBranches lists every branch in the worktree's stack.
	StackBranches(wt Worktree) ([]string, error)
	// AllResolved reports whether every branch in the stack is merged/closed.
	AllResolved(wt Worktree) (bool, error)
}

// NoneStacker is the no-op stacker: croft has no stack awareness.
type NoneStacker struct{}

// Sync does nothing.
func (NoneStacker) Sync(Worktree) (StackState, error) {
	return StackState{Detail: "no stacker provider"}, nil
}

// StackBranches returns no branches.
func (NoneStacker) StackBranches(Worktree) ([]string, error) { return nil, nil }

// AllResolved reports false: with no stacker, resolution is unknown, so the
// teardown gate must never fire automatically.
func (NoneStacker) AllResolved(Worktree) (bool, error) { return false, nil }

// ansiSeq matches ANSI color escape sequences in `gt` output.
var ansiSeq = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// branchToken matches a plausible branch-name token.
var branchToken = regexp.MustCompile(`^[A-Za-z0-9_./-]+$`)

// GraphiteStacker drives branch stacks with the Graphite CLI (`gt`).
type GraphiteStacker struct {
	bin string
}

// NewGraphiteStacker returns a Graphite-backed stacker. An empty bin resolves
// gt from PATH.
func NewGraphiteStacker(bin string) *GraphiteStacker {
	if bin == "" {
		bin = "gt"
	}
	return &GraphiteStacker{bin: bin}
}

// Sync rebases the worktree's stack onto the trunk. It runs non-interactively
// so it never blocks on a prompt; merged-branch teardown is croft's job, not
// gt's, so --force is deliberately omitted.
func (g *GraphiteStacker) Sync(wt Worktree) (StackState, error) {
	res, err := sh.Capture(g.bin, wt.Path, nil, "sync", "--no-interactive")
	if err != nil {
		return StackState{}, err
	}
	branches, _ := g.StackBranches(wt)
	return StackState{Branches: branches, Rebased: true, Detail: strings.TrimSpace(res)}, nil
}

// StackBranches lists the branches in the worktree's current stack, trunk
// excluded.
func (g *GraphiteStacker) StackBranches(wt Worktree) ([]string, error) {
	res, err := sh.Capture(g.bin, wt.Path, nil, "log", "short", "-s")
	if err != nil {
		return nil, err
	}
	return parseStackBranches(res), nil
}

// parseStackBranches extracts branch names from `gt log short -s` output:
// strip ANSI, then take the first branch-shaped token on each line.
func parseStackBranches(out string) []string {
	var branches []string
	for _, line := range strings.Split(out, "\n") {
		line = ansiSeq.ReplaceAllString(line, "")
		for _, tok := range strings.Fields(line) {
			if branchToken.MatchString(tok) && tok != "main" && tok != "master" {
				branches = append(branches, tok)
				break
			}
		}
	}
	return branches
}

// AllResolved reports whether every branch in the stack is in a terminal PR
// state (MERGED, CLOSED, or no PR). An empty/undeterminable stack reports
// false so the teardown gate never fires on uncertainty.
func (g *GraphiteStacker) AllResolved(wt Worktree) (bool, error) {
	branches, err := g.StackBranches(wt)
	if err != nil {
		return false, err
	}
	if len(branches) == 0 {
		return false, nil
	}
	states := loadPRStates(wt.Path)
	for _, br := range branches {
		switch states[br] {
		case "", "MERGED", "CLOSED":
			continue
		default:
			return false, nil
		}
	}
	return true, nil
}
