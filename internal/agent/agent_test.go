package agent

import (
	"regexp"
	"strings"
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
	"github.com/stretchr/testify/require"
)

var uuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestDeterministicSessionID(t *testing.T) {
	a := DeterministicSessionID("demo:feat:claude")
	b := DeterministicSessionID("demo:feat:claude")
	c := DeterministicSessionID("demo:other:claude")
	require.Equal(t, a, b, "session id should be stable")
	require.NotEqual(t, a, c, "different keys should produce different ids")
	require.Regexp(t, uuidV4, a, "session id should be a v4 UUID")
}

func TestClaudeLaunch(t *testing.T) {
	r := &ClaudeRunner{bin: "claude"}

	inv, err := r.Launch(Spec{Dir: "/wt", SessionID: "sid-1"})
	require.NoError(t, err)
	require.Equal(t, "claude", inv.Path)
	require.Equal(t, "/wt", inv.Dir)
	require.Contains(t, inv.Args, "--session-id")
	require.Contains(t, inv.Args, "sid-1")
	require.NotContains(t, inv.Args, "-p", "interactive launch should not be headless")

	headless, err := r.Launch(Spec{Headless: true})
	require.NoError(t, err)
	require.Contains(t, headless.Args, "-p")
	require.Contains(t, headless.Args, "--bare")
}

func TestClaudeResume(t *testing.T) {
	r := &ClaudeRunner{bin: "claude"}
	inv, err := r.Resume("sid-9", Spec{})
	require.NoError(t, err)
	require.Contains(t, inv.Args, "--resume")
	require.Contains(t, inv.Args, "sid-9")

	_, err = r.Resume("", Spec{})
	require.Error(t, err, "resume with an empty id should fail")
}

func TestCodexLaunch(t *testing.T) {
	r := &CodexRunner{bin: "codex"}

	inv, err := r.Launch(Spec{Headless: true, Profile: "dev", Prompt: "do it"})
	require.NoError(t, err)
	require.Subset(t, inv.Args, []string{"exec", "--json", "--profile", "dev"})

	interactive, err := r.Launch(Spec{})
	require.NoError(t, err)
	require.NotContains(t, interactive.Args, "exec", "interactive codex launch should not use exec")

	resume, err := r.Resume("th-1", Spec{Headless: true})
	require.NoError(t, err)
	require.Contains(t, resume.Args, "resume")
	require.Contains(t, resume.Args, "th-1")
}

func TestExecRunnerSubstitutes(t *testing.T) {
	r := &ExecRunner{name: "gemini", argv: []string{"gemini", "--cwd", "{dir}", "{prompt}"}}
	inv, err := r.Launch(Spec{Dir: "/wt/demo", Prompt: "fix bug"})
	require.NoError(t, err)
	require.Equal(t, "gemini", inv.Path)
	joined := strings.Join(inv.Args, " ")
	require.Contains(t, joined, "/wt/demo")
	require.Contains(t, joined, "fix bug")
}

func TestNewFactory(t *testing.T) {
	m := config.MachineConfig{}
	for _, runner := range []config.AgentRunner{config.RunnerClaude, config.RunnerCodex} {
		got, err := New(config.AgentConfig{Name: string(runner), Runner: runner}, m)
		require.NoError(t, err, "New(%s)", runner)
		require.Equal(t, string(runner), got.Name())
	}
	_, err := New(config.AgentConfig{Name: "x", Runner: config.RunnerExec}, m)
	require.Error(t, err, "exec runner without a command should fail")

	_, err = New(config.AgentConfig{Name: "x", Runner: "bogus"}, m)
	require.Error(t, err, "unknown runner should fail")
}
