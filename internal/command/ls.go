package command

import (
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
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

	// liveStatus shells out per worktree (rebase check, agent-window probe);
	// compute them concurrently so `ls` cost is one round-trip, not N.
	statuses := make([]string, len(slugs))
	var wg sync.WaitGroup
	for i, slug := range slugs {
		wg.Add(1)
		go func(i int, wt state.Worktree) {
			defer wg.Done()
			statuses[i] = ctx.liveStatus(wt)
		}(i, reg.Worktrees[slug])
	}
	wg.Wait()

	tw := tabwriter.NewWriter(out, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "SLUG\tBRANCH\tSTATUS\tPORTS\tPATH")
	for i, slug := range slugs {
		wt := reg.Worktrees[slug]
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			wt.Slug, wt.Branch, statuses[i], formatPorts(wt.Ports), wt.Path)
	}
	return tw.Flush()
}
