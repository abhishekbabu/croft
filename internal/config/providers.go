package config

import "strings"

// Provider and runner values are modeled as named string enums: a typed
// constant per value, an exhaustive list, and a Valid method. TOML decodes
// directly into these named string types.

// Multiplexer identifies a terminal-multiplexer backend.
type Multiplexer string

// Multiplexer values.
const (
	MultiplexerNone Multiplexer = "none"
	MultiplexerTmux Multiplexer = "tmux"
	MultiplexerCmux Multiplexer = "cmux"
)

// Multiplexers lists every valid multiplexer value.
var Multiplexers = []Multiplexer{MultiplexerNone, MultiplexerTmux, MultiplexerCmux}

// Valid reports whether m is a known multiplexer.
func (m Multiplexer) Valid() bool { return enumValid(m, Multiplexers) }

// InfraProvider identifies a container-stack backend.
type InfraProvider string

// InfraProvider values.
const (
	InfraNone          InfraProvider = "none"
	InfraDockerCompose InfraProvider = "docker-compose"
)

// InfraProviders lists every valid infra value.
var InfraProviders = []InfraProvider{InfraNone, InfraDockerCompose}

// Valid reports whether i is a known infra provider.
func (i InfraProvider) Valid() bool { return enumValid(i, InfraProviders) }

// Router identifies a routing backend.
type Router string

// Router values.
const (
	RouterNone     Router = "none"
	RouterPortless Router = "portless"
)

// Routers lists every valid router value.
var Routers = []Router{RouterNone, RouterPortless}

// Valid reports whether r is a known router.
func (r Router) Valid() bool { return enumValid(r, Routers) }

// Stacker identifies a branch-stack backend.
type Stacker string

// Stacker values.
const (
	StackerNone     Stacker = "none"
	StackerGraphite Stacker = "graphite"
)

// Stackers lists every valid stacker value.
var Stackers = []Stacker{StackerNone, StackerGraphite}

// Valid reports whether s is a known stacker.
func (s Stacker) Valid() bool { return enumValid(s, Stackers) }

// Coordination identifies an agent-coordination backend.
type Coordination string

// Coordination values.
const (
	CoordinationBasic            Coordination = "basic"
	CoordinationClaudeAgentTeams Coordination = "claude-agent-teams"
)

// Coordinations lists every valid coordination value.
var Coordinations = []Coordination{CoordinationBasic, CoordinationClaudeAgentTeams}

// Valid reports whether c is a known coordination backend.
func (c Coordination) Valid() bool { return enumValid(c, Coordinations) }

// AgentRunner identifies the runner that backs an agent.
type AgentRunner string

// AgentRunner values.
const (
	RunnerClaude AgentRunner = "claude"
	RunnerCodex  AgentRunner = "codex"
	RunnerExec   AgentRunner = "exec"
)

// AgentRunners lists every valid runner value.
var AgentRunners = []AgentRunner{RunnerClaude, RunnerCodex, RunnerExec}

// Valid reports whether r is a known agent runner.
func (r AgentRunner) Valid() bool { return enumValid(r, AgentRunners) }

// enumValid reports whether v is a member of set.
func enumValid[T comparable](v T, set []T) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

// enumList renders an enum's values as "a | b | c" for help text and errors.
func enumList[T ~string](set []T) string {
	parts := make([]string, len(set))
	for i, v := range set {
		parts[i] = string(v)
	}
	return strings.Join(parts, " | ")
}
