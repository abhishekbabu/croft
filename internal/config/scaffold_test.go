package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			Multiplexer: "tmux", Infra: "docker-compose",
			Router: "none", Stacker: "none", Coordination: "basic",
		},
		Agents: []AgentConfig{{Name: "claude", Runner: "claude"}},
		Hooks:  HooksSection{PostCreate: []string{"docker compose up -d"}},
	}
	p.applyDefaults()
	return p
}

func TestScaffoldRoundTrips(t *testing.T) {
	p := sampleProject()
	rendered := Scaffold(p)

	got, err := DecodeProject([]byte(rendered))
	if err != nil {
		t.Fatalf("scaffolded TOML failed to decode:\n%s\nerror: %v", rendered, err)
	}
	if got.Project.Name != p.Project.Name {
		t.Errorf("name round-trip: got %q want %q", got.Project.Name, p.Project.Name)
	}
	if got.Providers != p.Providers {
		t.Errorf("providers round-trip: got %+v want %+v", got.Providers, p.Providers)
	}
	if len(got.Agents) != 1 || got.Agents[0].Runner != "claude" {
		t.Errorf("agents round-trip: got %+v", got.Agents)
	}
	if got.Ports.Range != "3000-3999" {
		t.Errorf("ports round-trip: got %q", got.Ports.Range)
	}
}

func TestScaffoldEmptyCollectionsAreValidTOML(t *testing.T) {
	p := ProjectConfig{Project: ProjectSection{Name: "demo"}}
	p.applyDefaults()
	rendered := Scaffold(p)
	if !strings.Contains(rendered, "copy_files = []") {
		t.Errorf("empty slice should render as []:\n%s", rendered)
	}
	if _, err := DecodeProject([]byte(rendered)); err != nil {
		t.Fatalf("scaffold of minimal config must decode: %v", err)
	}
}

func TestExampleConfigIsValid(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "examples", "croft.toml"))
	if err != nil {
		t.Fatalf("read example config: %v", err)
	}
	if _, err := DecodeProject(data); err != nil {
		t.Fatalf("examples/croft.toml is invalid: %v", err)
	}
}
