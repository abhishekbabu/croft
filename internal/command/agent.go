package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/abhishekbabu/croft/internal/agent"
	"github.com/abhishekbabu/croft/internal/config"
	"github.com/abhishekbabu/croft/internal/provider"
	"github.com/abhishekbabu/croft/internal/state"
)

// findAgent looks up an agent declaration by name in the project config.
func findAgent(cfg config.ProjectConfig, name string) (config.AgentConfig, error) {
	for _, a := range cfg.Agents {
		if a.Name == name {
			return a, nil
		}
	}
	names := make([]string, len(cfg.Agents))
	for i, a := range cfg.Agents {
		names[i] = a.Name
	}
	avail := "none configured"
	if len(names) > 0 {
		avail = strings.Join(names, ", ")
	}
	return config.AgentConfig{}, fmt.Errorf("no agent %q in croft.toml (available: %s)", name, avail)
}

// launchAgent launches the named agent into the worktree's session and records
// the worktree as having a working agent.
func launchAgent(ctx *appContext, rec state.Worktree, agentName string, env map[string]string, out io.Writer) error {
	agentCfg, err := findAgent(ctx.Config, agentName)
	if err != nil {
		return err
	}
	runner, err := agent.New(agentCfg, ctx.Machine)
	if err != nil {
		return err
	}

	sessionID := agent.DeterministicSessionID(
		ctx.Config.Project.Name + ":" + rec.Slug + ":" + agentName)
	inv, err := runner.Launch(agent.Spec{
		Dir:       rec.Path,
		SessionID: sessionID,
		Env:       env,
	})
	if err != nil {
		return err
	}

	pw := ctx.providerWorktree(rec)
	session := provider.ProjectName(pw)
	if ctx.Providers.Multiplexer.Managed() && ctx.Providers.Multiplexer.HasWindow(session, "agent") {
		fmt.Fprintf(out, "Agent already running in worktree %q\n", rec.Slug)
		return nil
	}
	if err := ctx.Providers.Multiplexer.RunWindow(
		session, "agent", rec.Path, inv.Env, inv.Argv()); err != nil {
		return fmt.Errorf("launch agent: %w", err)
	}

	rec.Status = state.StatusWorking
	if err := ctx.Store.Put(rec); err != nil {
		return err
	}
	fmt.Fprintf(out, "Launched agent %q (%s) in worktree %q\n", agentName, runner.Name(), rec.Slug)
	return nil
}
