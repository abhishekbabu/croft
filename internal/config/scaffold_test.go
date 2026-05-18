package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func sampleProject() ProjectConfig {
	p := ProjectConfig{
		Project: ProjectSection{Name: "demo"},
		Worktree: WorktreeSection{
			DevCommand: "just dev",
			CopyFiles:  []string{".env.local"},
		},
		Ports: PortsSection{
			Range:    "3000-3999",
			Services: []string{"api", "postgres"},
		},
		Providers: ProvidersSection{
			Multiplexer: MultiplexerTmux, Infra: InfraDockerCompose,
			Router: RouterNone, Stacker: StackerNone, Coordination: CoordinationBasic,
		},
		Agents: []AgentConfig{{Name: "claude", Runner: RunnerClaude}},
		Hooks:  HooksSection{PostCreate: []string{"docker compose up -d"}},
	}
	p.applyDefaults()
	return p
}

func TestScaffoldRoundTrips(t *testing.T) {
	p := sampleProject()
	rendered := Scaffold(p)

	got, err := DecodeProject([]byte(rendered))
	require.NoError(t, err, "scaffolded TOML failed to decode:\n%s", rendered)
	require.Equal(t, p.Project.Name, got.Project.Name)
	require.Equal(t, p.Providers, got.Providers)
	require.Len(t, got.Agents, 1)
	require.Equal(t, RunnerClaude, got.Agents[0].Runner)
	require.Equal(t, "3000-3999", got.Ports.Range)
}

func TestScaffoldEmptyCollectionsAreValidTOML(t *testing.T) {
	p := ProjectConfig{Project: ProjectSection{Name: "demo"}}
	p.applyDefaults()
	rendered := Scaffold(p)
	require.Contains(t, rendered, "copy_files = []", "empty slice should render as []")

	_, err := DecodeProject([]byte(rendered))
	require.NoError(t, err, "scaffold of minimal config must decode")
}

func TestExampleConfigIsValid(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "examples", "croft.toml"))
	require.NoError(t, err)
	_, err = DecodeProject(data)
	require.NoError(t, err, "examples/croft.toml is invalid")
}
