package command

import (
	"path/filepath"

	"github.com/abhishekbabu/croft/internal/agent"
	"github.com/abhishekbabu/croft/internal/config"
	"github.com/abhishekbabu/croft/internal/provider"
)

// buildCoordination constructs the coordination backend for the project. It
// lives in the command layer rather than provider.New because the backend
// needs runtime wiring (peer state dir, agent launcher) that the provider
// package cannot assemble on its own.
//
// `claude-agent-teams` currently uses the same file-based backend as `basic`;
// native Agent Teams integration is deferred Phase 2 polish (PLAN.md §5).
func buildCoordination(ctx *appContext) provider.Coordination {
	switch ctx.Config.Providers.Coordination {
	case config.CoordinationBasic, config.CoordinationClaudeAgentTeams:
		peersDir := filepath.Join(ctx.Store.Dir(), "peers")
		session := ctx.Config.Project.Name + "-peers"
		return provider.NewBasicCoordination(peersDir, session, ctx.Providers.Multiplexer, agentLauncher(ctx))
	default:
		return provider.NoopCoordination{}
	}
}

// agentLauncher returns a provider.AgentLauncher that resolves a peer's agent
// against the project config and builds its launch invocation.
func agentLauncher(ctx *appContext) provider.AgentLauncher {
	return func(spec provider.PeerSpec) ([]string, map[string]string, error) {
		agentCfg, err := findAgent(ctx.Config, spec.Agent)
		if err != nil {
			return nil, nil, err
		}
		runner, err := agent.New(agentCfg, ctx.Machine)
		if err != nil {
			return nil, nil, err
		}
		env := map[string]string{"CROFT_PEER": spec.Name}
		if spec.Role != "" {
			env["CROFT_PEER_ROLE"] = spec.Role
		}
		inv, err := runner.Launch(agent.Spec{
			Dir:                spec.Dir,
			SessionID:          agent.DeterministicSessionID(ctx.Config.Project.Name + ":peer:" + spec.Name),
			AppendSystemPrompt: spec.Role,
			Env:                env,
		})
		if err != nil {
			return nil, nil, err
		}
		return inv.Argv(), inv.Env, nil
	}
}
