package config

import "fmt"

// Allowed provider values for each provider slot (see PLAN.md §6.3). The
// baseline / zero-dependency option is listed first in each set so it can
// double as the default.
var (
	// Multiplexers are the valid `providers.multiplexer` values. (cmux is a
	// planned backend but not yet implemented — it needs a dedicated adapter,
	// not the tmux CLI.)
	Multiplexers = []string{"none", "tmux"}
	// InfraProviders are the valid `providers.infra` values.
	InfraProviders = []string{"none", "docker-compose"}
	// Routers are the valid `providers.router` values.
	Routers = []string{"none", "portless"}
	// Stackers are the valid `providers.stacker` values.
	Stackers = []string{"none", "graphite"}
	// Coordinations are the valid `providers.coordination` values.
	Coordinations = []string{"basic", "claude-agent-teams"}
	// AgentRunners are the valid `agents[].runner` values.
	AgentRunners = []string{"claude", "codex", "exec"}
)

// contains reports whether set includes v.
func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

// validateEnum returns an error if value is not a member of set.
func validateEnum(field, value string, set []string) error {
	if !contains(set, value) {
		return fmt.Errorf("%s: %q is not one of %v", field, value, set)
	}
	return nil
}
