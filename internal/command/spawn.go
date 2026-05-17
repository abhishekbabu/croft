package command

import (
	"fmt"
	"io"
	"os"

	"github.com/abhishekbabu/croft/internal/provider"
	"github.com/spf13/cobra"
)

// NewSpawnCmd builds the `croft spawn` command, which spawns a named
// coordinated agent.
func NewSpawnCmd() *cobra.Command {
	var agentName, role, dir string
	cmd := &cobra.Command{
		Use:   "spawn <name> --agent <agent>",
		Short: "Spawn a named coordinated agent",
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
			if dir == "" {
				dir = cwd
			}
			return doSpawn(ctx, args[0], agentName, role, dir, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&agentName, "agent", "", "agent to run as this peer (required)")
	cmd.Flags().StringVar(&role, "role", "", "role text appended to the agent's system prompt")
	cmd.Flags().StringVar(&dir, "dir", "", "working directory for the peer (default: current directory)")
	return cmd
}

// doSpawn spawns a coordinated peer via the coordination provider.
func doSpawn(ctx *appContext, name, agentName, role, dir string, out io.Writer) error {
	if agentName == "" {
		return fmt.Errorf("--agent is required")
	}
	peer, err := buildCoordination(ctx).Spawn(provider.PeerSpec{
		Name:  name,
		Agent: agentName,
		Role:  role,
		Dir:   dir,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Spawned peer %q (agent %s, status %s)\n", peer.Name, peer.Agent, peer.Status)
	return nil
}
