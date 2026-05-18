// Command croft gives every git branch its own fully isolated, working
// development environment — checkout, container stack, ports, route, and
// optionally an AI coding agent.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/abhishekbabu/croft/internal/command"
	"github.com/abhishekbabu/croft/internal/sh"
	"github.com/spf13/cobra"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	// Cancel on SIGINT/SIGTERM so Ctrl-C terminates in-flight child processes:
	// every external command derives its context from sh's base context.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	sh.SetBaseContext(ctx)

	if err := newRootCmd().ExecuteContext(ctx); err != nil {
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

	root.AddCommand(
		command.NewInitCmd(),
		command.NewNewCmd(),
		command.NewLsCmd(),
		command.NewStatusCmd(),
		command.NewSyncCmd(),
		command.NewRmCmd(),
		command.NewDoctorCmd(),
		command.NewSpawnCmd(),
		command.NewFleetCmd(),
	)
	return root
}
