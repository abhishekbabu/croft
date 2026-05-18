package agent

import (
	"regexp"
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
