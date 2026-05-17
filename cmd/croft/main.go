// Command croft gives every git branch its own fully isolated, working
// development environment — checkout, container stack, ports, route, and
// optionally an AI coding agent.
package main

import (
	"fmt"
	"os"

	"github.com/abhishekbabu/croft/internal/command"
	"github.com/spf13/cobra"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "croft:", err)
		os.Exit(1)
	}
}

// newRootCmd builds the root `croft` command. Subcommands are attached here as
// they land milestone by milestone (see PLAN.md §10).
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "croft",
		Short: "Per-branch isolated development environments",
		Long: "croft gives every git branch its own fully isolated, working\n" +
			"development environment: checkout, container stack, ports, route,\n" +
			"and optionally an AI coding agent.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("croft {{.Version}}\n")

	root.AddCommand(command.NewInitCmd())
	return root
}
