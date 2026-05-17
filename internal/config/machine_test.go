package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMachineMissingFile(t *testing.T) {
	m, err := LoadMachineFrom(filepath.Join(t.TempDir(), "absent.toml"))
	if err != nil {
		t.Fatalf("missing machine config should not error: %v", err)
	}
	if len(m.Bins) != 0 || m.Defaults.Agent != "" {
		t.Errorf("missing machine config should be zero-value, got %+v", m)
	}
}

func TestLoadMachineFrom(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, MachineFileName)
	src := `
[bins]
claude = "/opt/homebrew/bin/claude"

[defaults]
agent = "codex"
effort = "xhigh"

[aws]
sso_profile = "demo-dev"
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := LoadMachineFrom(path)
	if err != nil {
		t.Fatalf("LoadMachineFrom: %v", err)
	}
	if m.Bin("claude") != "/opt/homebrew/bin/claude" {
		t.Errorf("Bin(claude) = %q", m.Bin("claude"))
	}
	if m.Bin("codex") != "codex" {
		t.Errorf("Bin(codex) unset should fall back to %q, got %q", "codex", m.Bin("codex"))
	}
	if m.Defaults.Effort != "xhigh" {
		t.Errorf("defaults.effort = %q, want xhigh", m.Defaults.Effort)
	}
	if m.AWS.SSOProfile != "demo-dev" {
		t.Errorf("aws.sso_profile = %q, want demo-dev", m.AWS.SSOProfile)
	}
}

func TestDefaultAgent(t *testing.T) {
	p := ProjectConfig{Agents: []AgentConfig{{Name: "claude", Runner: "claude"}}}

	if got := (MachineConfig{}).DefaultAgent(p); got != "claude" {
		t.Errorf("DefaultAgent with no override = %q, want claude (first agent)", got)
	}
	m := MachineConfig{Defaults: DefaultsSection{Agent: "codex"}}
	if got := m.DefaultAgent(p); got != "codex" {
		t.Errorf("DefaultAgent with override = %q, want codex", got)
	}
	if got := (MachineConfig{}).DefaultAgent(ProjectConfig{}); got != "" {
		t.Errorf("DefaultAgent with nothing configured = %q, want empty", got)
	}
}
