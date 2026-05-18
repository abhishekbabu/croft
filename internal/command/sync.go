package command

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/abhishekbabu/croft/internal/state"
	"github.com/abhishekbabu/croft/internal/worktree"
	"github.com/spf13/cobra"
)

// NewSyncCmd builds the `croft sync` command, which rebases worktree branch
// stacks against the trunk.
func NewSyncCmd() *cobra.Command {
	var prune bool
	cmd := &cobra.Command{
		Use:   "sync [branch]",
		Short: "Sync worktree branch stacks against the trunk",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCwd()
			if err != nil {
				return err
			}
			branch := ""
			if len(args) == 1 {
				branch = args[0]
			}
			return doSync(ctx, branch, prune, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&prune, "prune", false, "tear down worktrees whose branch stack is fully resolved")
	return cmd
}

// doSync syncs one worktree (when branch is set) or every registered worktree.
func doSync(ctx *appContext, branch string, prune bool, out io.Writer) error {
	reg, err := ctx.Store.Load()
	if err != nil {
		return err
	}

	var targets []state.Worktree
	if branch != "" {
		rec, ok := reg.Worktrees[worktree.Slugify(branch)]
		if !ok {
			return fmt.Errorf("no worktree %q", branch)
		}
		targets = []state.Worktree{rec}
	} else {
		slugs := make([]string, 0, len(reg.Worktrees))
		for s := range reg.Worktrees {
			slugs = append(slugs, s)
		}
		sort.Strings(slugs)
		for _, s := range slugs {
			targets = append(targets, reg.Worktrees[s])
		}
	}
	if len(targets) == 0 {
		fmt.Fprintln(out, "No worktrees to sync.")
		return nil
	}

	// Unlike ls/doctor, sync is deliberately sequential. It auto-stashes each
	// worktree, and refs/stash (with its stash stack) is shared across all
	// worktrees of a repo — concurrent stash/pop would interleave entries, so
	// one worktree's pop could restore another's changes. With the Graphite
	// stacker it is worse: `gt sync` rebases the whole stack and trunk, not
	// just the checked-out branch, contending on shared ref locks, and
	// Graphite is not built for concurrent invocation on one repo. Sequential
	// execution also keeps per-worktree progress output from interleaving.
	var failures int
	for _, rec := range targets {
		if err := syncOne(ctx, rec, prune, out); err != nil {
			fmt.Fprintf(out, "  x %s\n", err)
			failures++
		}
	}
	if failures > 0 {
		return fmt.Errorf("%d worktree(s) failed to sync", failures)
	}
	return nil
}

// syncOne syncs a single worktree: refuse mid-rebase, auto-stash, rebase the
// stack, pop the stash, then apply the stack-resolved teardown gate.
func syncOne(ctx *appContext, rec state.Worktree, prune bool, out io.Writer) error {
	fmt.Fprintf(out, "%s (%s)\n", rec.Slug, rec.Branch)

	if !dirExists(rec.Path) {
		return fmt.Errorf("worktree directory missing: %s", rec.Path)
	}
	if ctx.Manager.InRebase(rec.Path) {
		return fmt.Errorf("%s is mid-rebase — finish with `gt continue` or `git rebase --abort`", rec.Slug)
	}

	stashed, err := ctx.Manager.Stash(rec.Path, "croft sync: auto-stash")
	if err != nil {
		return fmt.Errorf("auto-stash %s: %w", rec.Slug, err)
	}

	pw := ctx.providerWorktree(rec)
	st, syncErr := ctx.Providers.Stacker.Sync(pw)

	if stashed {
		if popErr := ctx.Manager.StashPop(rec.Path); popErr != nil {
			// Silent stash loss is worse than any sync error — surface loudly.
			return fmt.Errorf("FAILED to pop stash in %s — your changes are safe in stash@{0}; "+
				"recover with `git -C %s stash pop`", rec.Path, rec.Path)
		}
	}
	if syncErr != nil {
		return fmt.Errorf("sync %s: %w", rec.Slug, syncErr)
	}

	if detail := firstLine(st.Detail); detail != "" {
		fmt.Fprintf(out, "  synced: %s\n", detail)
	} else {
		fmt.Fprintln(out, "  synced")
	}

	resolved, _ := ctx.Providers.Stacker.AllResolved(pw)
	if resolved {
		if prune {
			fmt.Fprintln(out, "  stack fully resolved — tearing down")
			return doRm(ctx, rec.Branch, false, out)
		}
		fmt.Fprintf(out, "  stack fully resolved — tear down with: croft rm %s\n", rec.Branch)
	}
	return nil
}

// firstLine returns the first non-empty line of s, trimmed.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}
