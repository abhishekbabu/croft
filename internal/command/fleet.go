package command

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/abhishekbabu/croft/internal/provider"
	"github.com/spf13/cobra"
)

// NewFleetCmd builds the `croft fleet` command group for cross-worktree peer
// status and dispatch.
func NewFleetCmd() *cobra.Command {
	fleet := &cobra.Command{
		Use:   "fleet",
		Short: "Cross-worktree agent coordination",
		Args:  cobra.NoArgs,
	}
	fleet.AddCommand(fleetStatusCmd(), fleetMsgCmd())
	return fleet
}

// fleetStatusCmd builds `croft fleet status`.
func fleetStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "List coordinated peers across all worktrees",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := contextFromCwd()
			if err != nil {
				return err
			}
			return doFleetStatus(ctx, cmd.OutOrStdout())
		},
	}
}

// fleetMsgCmd builds `croft fleet msg`.
func fleetMsgCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "msg <peer> <message>",
		Short: "Send a message to a peer's mailbox",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCwd()
			if err != nil {
				return err
			}
			return doFleetMsg(ctx, args[0], strings.Join(args[1:], " "), cmd.OutOrStdout())
		},
	}
}

// doFleetStatus prints every coordinated peer as a table.
func doFleetStatus(ctx *appContext, out io.Writer) error {
	peers, err := buildCoordination(ctx).Status()
	if err != nil {
		return err
	}
	if len(peers) == 0 {
		fmt.Fprintln(out, "No peers. Spawn one with `croft spawn <name> --agent <agent>`.")
		return nil
	}
	tw := tabwriter.NewWriter(out, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "PEER\tAGENT\tSTATUS")
	for _, p := range peers {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", p.Name, p.Agent, p.Status)
	}
	return tw.Flush()
}

// doFleetMsg dispatches a message to a peer.
func doFleetMsg(ctx *appContext, peer, msg string, out io.Writer) error {
	if err := buildCoordination(ctx).Dispatch(provider.Peer{Name: peer}, msg); err != nil {
		return err
	}
	fmt.Fprintf(out, "Delivered message to %q\n", peer)
	return nil
}
