package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadMachineMissingFile(t *testing.T) {
	m, err := LoadMachineFrom(filepath.Join(t.TempDir(), "absent.toml"))
	require.NoError(t, err, "missing machine config should not error")
	require.Empty(t, m.Bins)
	require.Empty(t, m.Defaults.Agent)
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
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	m, err := LoadMachineFrom(path)
	require.NoError(t, err)
	require.Equal(t, "/opt/homebrew/bin/claude", m.Bin("claude"))
	require.Equal(t, "codex", m.Bin("codex"), "unset bin should fall back to the tool name")
	require.Equal(t, "xhigh", m.Defaults.Effort)
	require.Equal(t, "demo-dev", m.AWS.SSOProfile)
}

func TestLoadMachineViaXDG(t *testing.T) {
	cfgHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	path, err := MachineConfigPath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(cfgHome, "croft", MachineFileName), path)

	// No file there yet: LoadMachine returns a zero config without error.
	m, err := LoadMachine()
	require.NoError(t, err)
	require.Empty(t, m.Bins)

	// Once a config exists at the resolved path, LoadMachine reads it.
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("[defaults]\nagent = \"claude\"\n"), 0o644))
	m, err = LoadMachine()
	require.NoError(t, err)
	require.Equal(t, "claude", m.Defaults.Agent)
}

func TestDefaultAgent(t *testing.T) {
	p := ProjectConfig{Agents: []AgentConfig{{Name: "claude", Runner: RunnerClaude}}}

	require.Equal(t, "claude", MachineConfig{}.DefaultAgent(p), "no override falls back to the first agent")

	m := MachineConfig{Defaults: DefaultsSection{Agent: "codex"}}
	require.Equal(t, "codex", m.DefaultAgent(p), "machine override wins")

	require.Empty(t, MachineConfig{}.DefaultAgent(ProjectConfig{}), "nothing configured yields empty")
}
