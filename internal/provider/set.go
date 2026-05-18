package provider

import (
	"fmt"

	"github.com/abhishekbabu/croft/internal/config"
	"github.com/abhishekbabu/croft/internal/env"
)

// Set is the bundle of provider implementations selected by a project's
// configuration. Core code holds a Set and calls the interfaces; it never
// branches on which backend is active.
type Set struct {
	Multiplexer  Multiplexer
	Infra        Infra
	Router       Router
	Stacker      Stacker
	Coordination Coordination
}

// New builds a Set from the project's provider configuration and the machine
// config (for binary paths). stateDir is where stateful providers persist
// their bookkeeping. Unselected or not-yet-implemented providers fall back to
// their no-op implementation, so core always has a usable Set.
func New(p config.ProvidersSection, m config.MachineConfig, stateDir string) (Set, error) {
	s := Set{
		Multiplexer:  NoneMultiplexer{},
		Infra:        NoneInfra{},
		Router:       NoneRouter{},
		Stacker:      NoneStacker{},
		Coordination: NoopCoordination{},
	}

	switch p.Multiplexer {
	case "", config.MultiplexerNone:
	case config.MultiplexerTmux:
		s.Multiplexer = NewTmuxMultiplexer(m.Bin("tmux"))
	case config.MultiplexerCmux:
		// cmux can only drive surfaces from inside a cmux terminal; fail fast
		// here rather than partway through a command.
		if env.CmuxSurfaceID() == "" {
			return Set{}, fmt.Errorf(`providers.multiplexer is %q but croft is not `+
				`running inside a cmux terminal ($CMUX_SURFACE_ID is unset)`, config.MultiplexerCmux)
		}
		s.Multiplexer = NewCmuxMultiplexer(m.Bin("cmux"), stateDir)
	default:
		return Set{}, fmt.Errorf("unsupported multiplexer provider %q", p.Multiplexer)
	}

	switch p.Infra {
	case "", config.InfraNone:
	case config.InfraDockerCompose:
		s.Infra = NewComposeInfra(m.Bin("docker"))
	default:
		return Set{}, fmt.Errorf("unknown infra provider %q", p.Infra)
	}

	switch p.Router {
	case "", config.RouterNone:
	case config.RouterPortless:
		s.Router = NewPortlessRouter(m.Bin("portless"))
	default:
		return Set{}, fmt.Errorf("unknown router provider %q", p.Router)
	}

	switch p.Stacker {
	case "", config.StackerNone:
	case config.StackerGraphite:
		s.Stacker = NewGraphiteStacker(m.Bin("gt"))
	default:
		return Set{}, fmt.Errorf("unknown stacker provider %q", p.Stacker)
	}

	switch p.Coordination {
	case "", config.CoordinationBasic, config.CoordinationClaudeAgentTeams:
		// The basic backend is wired in the command layer (buildCoordination).
	default:
		return Set{}, fmt.Errorf("unknown coordination provider %q", p.Coordination)
	}

	return s, nil
}
