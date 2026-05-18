package provider

import (
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
	"github.com/abhishekbabu/croft/internal/sh"
	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	env := Env(Worktree{
		Slug:  "feat",
		Path:  "/wt/demo.feat",
		Ports: map[string]int{"api": 3000, "postgres": 3001},
	})
	require.Equal(t, "feat", env["CROFT_SLUG"])
	require.Equal(t, "croft-feat", env["COMPOSE_PROJECT_NAME"])
	require.Equal(t, "3000", env["API_PORT"])
	require.Equal(t, "3001", env["POSTGRES_PORT"])
}

func TestNewSelectsImplementations(t *testing.T) {
	set, err := New(config.ProvidersSection{
		Multiplexer:  config.MultiplexerTmux,
		Infra:        config.InfraDockerCompose,
		Router:       config.RouterNone,
		Stacker:      config.StackerNone,
		Coordination: config.CoordinationBasic,
	}, config.MachineConfig{}, "")
	require.NoError(t, err)
	require.IsType(t, &TmuxMultiplexer{}, set.Multiplexer)
	require.IsType(t, &ComposeInfra{}, set.Infra)
}

func TestNewDefaultsToNoOp(t *testing.T) {
	set, err := New(config.ProvidersSection{}, config.MachineConfig{}, "")
	require.NoError(t, err)
	require.IsType(t, NoneMultiplexer{}, set.Multiplexer)
	require.IsType(t, NoneInfra{}, set.Infra)
}

func TestNewRejectsUnknownProvider(t *testing.T) {
	_, err := New(config.ProvidersSection{Infra: "nomad"}, config.MachineConfig{}, "")
	require.Error(t, err, "unknown infra provider should fail")
}

func TestNoneProvidersAreInert(t *testing.T) {
	wt := Worktree{Slug: "x", Path: "/tmp/x"}
	require.NoError(t, NoneInfra{}.Up(wt))

	st, err := NoneInfra{}.Status(wt)
	require.NoError(t, err)
	require.False(t, st.Up, "NoneInfra.Status should report down")

	require.NoError(t, NoneMultiplexer{}.CreateSession("x", "/tmp", nil))

	url, err := NoneRouter{}.Register(wt)
	require.NoError(t, err)
	require.Empty(t, url, "NoneRouter.Register should yield no URL")

	ok, err := NoneStacker{}.AllResolved(wt)
	require.NoError(t, err)
	require.False(t, ok, "NoneStacker.AllResolved must be false")
}

func TestTmuxLifecycle(t *testing.T) {
	if !sh.Look("tmux") {
		t.Skip("tmux not installed")
	}
	tm := NewTmuxMultiplexer("")
	const name = "croft-test-session"
	_ = tm.Kill(name) // clean any leftover
	t.Cleanup(func() { _ = tm.Kill(name) })

	require.NoError(t, tm.CreateSession(name, t.TempDir(), map[string]string{"FOO": "bar"}))
	require.NoError(t, tm.CreateSession(name, t.TempDir(), nil), "CreateSession is idempotent")
	require.True(t, tm.hasSession(name), "session should exist after CreateSession")

	require.NoError(t, tm.Kill(name))
	require.NoError(t, tm.Kill(name), "Kill is idempotent")
}
