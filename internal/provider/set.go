package provider

import (
	"fmt"

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
// config (for binary paths). Unselected or not-yet-implemented providers fall
// back to their no-op implementation, so core always has a usable Set.
func New(p config.ProvidersSection, m config.MachineConfig) (Set, error) {
	s := Set{
		Multiplexer:  NoneMultiplexer{},
		Infra:        NoneInfra{},
		Router:       NoneRouter{},
		Stacker:      NoneStacker{},
		Coordination: NoopCoordination{},
	}

	switch p.Multiplexer {
	case "", "none":
	case "tmux", "cmux":
		// cmux shares the tmux control protocol; a dedicated adapter is
		// planned for a later milestone.
		s.Multiplexer = NewTmuxMultiplexer(m.Bin(p.Multiplexer))
	default:
		return Set{}, fmt.Errorf("unknown multiplexer provider %q", p.Multiplexer)
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
		// Implemented in a later milestone; no-op until then.
	default:
		return Set{}, fmt.Errorf("unknown router provider %q", p.Router)
	}

	switch p.Stacker {
	case "", "none":
	case "graphite":
		// Implemented in a later milestone; no-op until then.
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
