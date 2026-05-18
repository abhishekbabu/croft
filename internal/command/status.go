package command

import (
	"fmt"
	"io"
	"os"

	"github.com/abhishekbabu/croft/internal/worktree"
	"github.com/spf13/cobra"
)

// NewStatusCmd builds the `croft status` command, which shows detail for one
// worktree.
func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <branch>",
		Short: "Show detail for one worktree",
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
			return doStatus(ctx, args[0], cmd.OutOrStdout())
		},
	}
}

// doStatus prints the registry detail for the worktree named by branch.
func doStatus(ctx *appContext, branch string, out io.Writer) error {
	slug := worktree.Slugify(branch)
	wt, found, err := ctx.Store.Get(slug)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("no worktree %q (looked up slug %q)", branch, slug)
	}

	status := displayStatus(ctx.liveStatus(wt))
	fmt.Fprintf(out, "slug:    %s\n", wt.Slug)
	fmt.Fprintf(out, "branch:  %s\n", wt.Branch)
	fmt.Fprintf(out, "path:    %s\n", wt.Path)
	if wt.URL != "" {
		fmt.Fprintf(out, "url:     %s\n", wt.URL)
	}
	fmt.Fprintf(out, "exists:  %t\n", dirExists(wt.Path))
	fmt.Fprintf(out, "status:  %s\n", status)
	fmt.Fprintf(out, "ports:   %s\n", formatPorts(wt.Ports))
	fmt.Fprintf(out, "created: %s\n", wt.Created.Format("2006-01-02 15:04:05 MST"))
	return nil
}
