package command

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/abhishekbabu/croft/internal/state"
	"github.com/abhishekbabu/croft/internal/worktree"
	"github.com/spf13/cobra"
)

// NewNewCmd builds the `croft new` command, which creates an isolated worktree
// for a branch.
func NewNewCmd() *cobra.Command {
	var from string
	cmd := &cobra.Command{
		Use:   "new <branch>",
		Short: "Create an isolated worktree for a branch",
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
			return doNew(ctx, args[0], from, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "start point for a new branch (default: current HEAD)")
	return cmd
}

// doNew creates (or idempotently reports) the worktree for branch.
func doNew(ctx *appContext, branch, from string, out io.Writer) error {
	slug := worktree.Slugify(branch)
	if slug == "" {
		return fmt.Errorf("branch %q produces an empty slug", branch)
	}

	reg, err := ctx.Store.Load()
	if err != nil {
		return err
	}
	if existing, ok := reg.Worktrees[slug]; ok && dirExists(existing.Path) {
		fmt.Fprintf(out, "Worktree %q already exists at %s\n", slug, existing.Path)
		return nil
	}

	id := worktree.Resolve(slug, ctx.WorktreeRoot, ctx.Config.Worktree.Naming)
	if err := os.MkdirAll(ctx.WorktreeRoot, 0o755); err != nil {
		return fmt.Errorf("create worktree root: %w", err)
	}

	ports := map[string]int{}
	if ctx.Config.Ports.Range != "" && len(ctx.Config.Ports.Services) > 0 {
		low, high, err := ctx.Config.Ports.Bounds()
		if err != nil {
			return err
		}
		ports, err = worktree.AllocatePorts(low, high, ctx.Config.Ports.Services, takenPorts(reg))
		if err != nil {
			return err
		}
	}

	if err := ctx.Manager.Add(id.Path, branch, from); err != nil {
		return err
	}
	rec := state.Worktree{
		Slug:    slug,
		Branch:  branch,
		Path:    id.Path,
		Ports:   ports,
		Created: time.Now().UTC(),
	}
	if err := ctx.Store.Put(rec); err != nil {
		// Roll the worktree back so a failed `new` leaves nothing behind.
		_ = ctx.Manager.Remove(id.Path, true)
		return fmt.Errorf("record worktree (rolled back): %w", err)
	}

	fmt.Fprintf(out, "Created worktree %q\n", slug)
	fmt.Fprintf(out, "  branch: %s\n", branch)
	fmt.Fprintf(out, "  path:   %s\n", id.Path)
	if len(ports) > 0 {
		fmt.Fprintf(out, "  ports:  %s\n", formatPorts(ports))
	}
	return nil
}
