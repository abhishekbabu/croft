package command

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/abhishekbabu/croft/internal/state"
	"github.com/spf13/cobra"
)

// NewLsCmd builds the `croft ls` command, which lists croft-managed worktrees.
func NewLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List croft-managed worktrees",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(cwd)
			if err != nil {
				return err
			}
			return doLs(ctx, cmd.OutOrStdout())
		},
	}
}

// doLs prints the worktree registry as a table.
func doLs(ctx *appContext, out io.Writer) error {
	reg, err := ctx.Store.Load()
	if err != nil {
		return err
	}
	if len(reg.Worktrees) == 0 {
		fmt.Fprintln(out, "No worktrees yet. Create one with `croft new <branch>`.")
		return nil
	}

	slugs := make([]string, 0, len(reg.Worktrees))
	for slug := range reg.Worktrees {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	tw := tabwriter.NewWriter(out, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "SLUG\tBRANCH\tSTATUS\tPORTS\tPATH")
	for _, slug := range slugs {
		wt := reg.Worktrees[slug]
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			wt.Slug, wt.Branch, healthLabel(wt), formatPorts(wt.Ports), wt.Path)
	}
	return tw.Flush()
}

// healthLabel summarizes a worktree's state for the listing.
func healthLabel(wt state.Worktree) string {
	if !dirExists(wt.Path) {
		return "missing"
	}
	if wt.Status == "" {
		return "-"
	}
	return wt.Status
}
