package command

import (
	"fmt"
	"io"
	"os"

	"github.com/abhishekbabu/croft/internal/provider"
	"github.com/abhishekbabu/croft/internal/state"
	"github.com/abhishekbabu/croft/internal/worktree"
	"github.com/spf13/cobra"
)

// NewRmCmd builds the `croft rm` command, which tears a worktree's environment
// down.
func NewRmCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "rm <branch>",
		Short: "Remove a worktree and tear down its environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(cwd)
			if err != nil {
				return err
			}
			return doRm(ctx, args[0], force, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "remove even when the worktree has uncommitted changes")
	return cmd
}

// doRm tears down the worktree named by branch: pre_remove hooks, container
// stack, session, then the worktree itself. It is idempotent — a worktree that
// is already partly gone is cleaned up the rest of the way without error.
func doRm(ctx *appContext, branch string, force bool, out io.Writer) error {
	slug := worktree.Slugify(branch)

	rec, found, err := ctx.Store.Get(slug)
	if err != nil {
		return err
	}
	path := rec.Path
	if !found {
		// Not registered — fall back to the resolved path so a leftover
		// directory can still be cleaned up.
		rec = state.Worktree{Slug: slug}
		path = worktree.Resolve(slug, ctx.WorktreeRoot, ctx.Config.Worktree.Naming).Path
		rec.Path = path
	}

	if !found && !dirExists(path) {
		fmt.Fprintf(out, "Nothing to remove for %q\n", slug)
		return nil
	}

	pw := ctx.providerWorktree(rec)
	env := provider.Env(pw)

	// Tear the environment down before removing the checkout, since hooks and
	// the compose file live inside it.
	if dirExists(path) {
		if err := runHooks("pre_remove", ctx.Config.Hooks.PreRemove, path, env, out); err != nil {
			return err
		}
	}
	if err := ctx.Providers.Infra.Down(pw); err != nil {
		return fmt.Errorf("bring infra down: %w", err)
	}
	if err := ctx.Providers.Multiplexer.Kill(provider.ProjectName(pw)); err != nil {
		return fmt.Errorf("kill session: %w", err)
	}

	if dirExists(path) {
		if err := ctx.Manager.Remove(path, force); err != nil {
			return err
		}
	}
	_ = ctx.Manager.Prune()
	if err := ctx.Store.Delete(slug); err != nil {
		return err
	}

	fmt.Fprintf(out, "Removed worktree %q\n", slug)
	return nil
}
