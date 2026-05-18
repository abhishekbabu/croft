package command

import (
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// execArgs runs a freshly built command with args, discarding its output, and
// returns the error. Argument-count validation happens before RunE, so a
// wrong-arity invocation fails here without ever touching the filesystem.
func execArgs(cmd *cobra.Command, args ...string) error {
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd.Execute()
}

func TestCommandArgValidation(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() *cobra.Command
		args []string
	}{
		{"init rejects args", NewInitCmd, []string{"extra"}},
		{"ls rejects args", NewLsCmd, []string{"extra"}},
		{"doctor rejects args", NewDoctorCmd, []string{"extra"}},
		{"status needs a branch", NewStatusCmd, nil},
		{"status rejects two args", NewStatusCmd, []string{"a", "b"}},
		{"new needs a branch", NewNewCmd, nil},
		{"new rejects two args", NewNewCmd, []string{"a", "b"}},
		{"rm needs a branch", NewRmCmd, nil},
		{"spawn needs a name", NewSpawnCmd, nil},
		{"sync rejects two args", NewSyncCmd, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, execArgs(tt.cmd(), tt.args...),
				"%s: expected an argument-validation error", tt.name)
		})
	}
}

func TestCommandFlagsRegistered(t *testing.T) {
	tests := []struct {
		name  string
		cmd   *cobra.Command
		flags []string
	}{
		{"init", NewInitCmd(), []string{"force", "yes"}},
		{"doctor", NewDoctorCmd(), []string{"fix"}},
		{"sync", NewSyncCmd(), []string{"prune"}},
		{"rm", NewRmCmd(), []string{"force"}},
		{"new", NewNewCmd(), []string{"from", "agent"}},
		{"spawn", NewSpawnCmd(), []string{"agent", "role", "dir"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.flags {
				require.NotNil(t, tt.cmd.Flags().Lookup(f),
					"%s should register --%s", tt.name, f)
			}
		})
	}
}

func TestFleetCmdWiring(t *testing.T) {
	fleet := NewFleetCmd()
	require.True(t, fleet.HasSubCommands(), "fleet should have subcommands")

	// `fleet msg` requires at least a slug and a peer.
	require.Error(t, execArgs(NewFleetCmd(), "msg", "only-one-arg"),
		"fleet msg with one arg should fail validation")
}
