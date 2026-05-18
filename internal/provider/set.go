package provider

import (
	"fmt"
	"os"

	"github.com/abhishekbabu/croft/internal/config"
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
	case "", "none":
	case "tmux":
		s.Multiplexer = NewTmuxMultiplexer(m.Bin("tmux"))
	case "cmux":
		// cmux can only drive surfaces from inside a cmux terminal; fail fast
		// here rather than partway through a command.
		if os.Getenv("CMUX_SURFACE_ID") == "" {
			return Set{}, fmt.Errorf(`providers.multiplexer is "cmux" but croft is not ` +
				`running inside a cmux terminal ($CMUX_SURFACE_ID is unset)`)
		}
		s.Multiplexer = NewCmuxMultiplexer(m.Bin("cmux"), stateDir)
	default:
		return Set{}, fmt.Errorf("unsupported multiplexer provider %q", p.Multiplexer)
	}

	switch p.Infra {
	case "", "none":
	case "docker-compose":
		s.Infra = NewComposeInfra(m.Bin("docker"))
	default:
		return Set{}, fmt.Errorf("unknown infra provider %q", p.Infra)
	}

	switch p.Router {
	case "", "none":
	case "portless":
		s.Router = NewPortlessRouter(m.Bin("portless"))
	default:
		return Set{}, fmt.Errorf("unknown router provider %q", p.Router)
	}

	switch p.Stacker {
	case "", "none":
	case "graphite":
		s.Stacker = NewGraphiteStacker(m.Bin("gt"))
	default:
		return Set{}, fmt.Errorf("unknown stacker provider %q", p.Stacker)
	}

	switch p.Coordination {
	case "", "basic", "claude-agent-teams":
		// Implemented in a later milestone; placeholder until then.
	default:
		return Set{}, fmt.Errorf("unknown coordination provider %q", p.Coordination)
	}

	return s, nil
}
