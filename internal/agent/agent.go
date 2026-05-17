// Package agent builds the process invocations for the AI coding agents croft
// can launch into a worktree — Claude, Codex, or any generic CLI agent. It
// never reimplements an agent; it only constructs how to shell out (PLAN.md
// §6.4).
package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/abhishekbabu/croft/internal/config"
)

// Spec describes a request to launch or resume an agent.
type Spec struct {
	Dir                string            // working directory (the worktree)
	SessionID          string            // deterministic id to pin the session, when supported
	Prompt             string            // initial prompt/task (optional)
	AppendSystemPrompt string            // dynamic role text (optional)
	Model              string            // model override (optional)
	Profile            string            // codex profile name (optional)
	Headless           bool              // scripted (-p / exec --json) vs interactive
	Env                map[string]string // environment exported to the agent
}

// Invocation is a fully resolved process to run.
type Invocation struct {
	Path string
	Args []string
	Env  map[string]string
	Dir  string
}

// Argv returns the invocation as a single argument vector.
func (inv Invocation) Argv() []string {
	return append([]string{inv.Path}, inv.Args...)
}

// Runner builds the process invocation for an agent CLI.
type Runner interface {
	// Name identifies the runner (claude, codex, exec).
	Name() string
	// Launch builds the invocation that starts a new agent session.
	Launch(spec Spec) (Invocation, error)
	// Resume builds the invocation that resumes an existing session.
	Resume(sessionID string, spec Spec) (Invocation, error)
}

// New builds the Runner for an agent configuration, resolving binary paths
// from the machine config.
func New(a config.AgentConfig, m config.MachineConfig) (Runner, error) {
	switch a.Runner {
	case "claude":
		return &ClaudeRunner{bin: m.Bin("claude")}, nil
	case "codex":
		return &CodexRunner{bin: m.Bin("codex")}, nil
	case "exec":
		if len(a.Command) == 0 {
			return nil, fmt.Errorf("exec runner %q has no command", a.Name)
		}
		return &ExecRunner{name: a.Name, argv: a.Command}, nil
	default:
		return nil, fmt.Errorf("unknown agent runner %q", a.Runner)
	}
}

// DeterministicSessionID derives a stable, RFC-4122-shaped (version 4) UUID
// from key, so the same worktree+agent always pins the same agent session.
func DeterministicSessionID(key string) string {
	sum := sha256.Sum256([]byte("croft:" + key))
	b := sum[:16]
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	h := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}
