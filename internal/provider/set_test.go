package provider

import (
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
	"github.com/stretchr/testify/require"
)

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

func TestNewSelectsStackerAndRouter(t *testing.T) {
	set, err := New(config.ProvidersSection{
		Router:       config.RouterPortless,
		Stacker:      config.StackerGraphite,
		Coordination: config.CoordinationBasic,
	}, config.MachineConfig{}, "")
	require.NoError(t, err)
	require.IsType(t, &PortlessRouter{}, set.Router)
	require.IsType(t, &GraphiteStacker{}, set.Stacker)
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

func TestNewSelectsCmux(t *testing.T) {
	t.Setenv("CMUX_SURFACE_ID", "test-surface")
	set, err := New(config.ProvidersSection{Multiplexer: config.MultiplexerCmux}, config.MachineConfig{}, t.TempDir())
	require.NoError(t, err)
	require.IsType(t, &CmuxMultiplexer{}, set.Multiplexer)
}

func TestNewRejectsCmuxOutsideCmux(t *testing.T) {
	t.Setenv("CMUX_SURFACE_ID", "")
	_, err := New(config.ProvidersSection{Multiplexer: config.MultiplexerCmux}, config.MachineConfig{}, t.TempDir())
	require.Error(t, err, "multiplexer=cmux outside a cmux terminal should fail")
}
