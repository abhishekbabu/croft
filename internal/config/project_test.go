package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeProjectAppliesDefaults(t *testing.T) {
	p, err := DecodeProject([]byte(`
[project]
name = "demo"
`))
	require.NoError(t, err)
	require.Equal(t, "../worktrees", p.Worktree.Root)
	require.Equal(t, "demo.{slug}", p.Worktree.Naming)
	require.Equal(t, MultiplexerNone, p.Providers.Multiplexer)
	require.Equal(t, CoordinationBasic, p.Providers.Coordination)
}

func TestDecodeProjectValidationErrors(t *testing.T) {
	cases := map[string]string{
		"missing name": `
[worktree]
naming = "{slug}"
`,
		"naming without slug": `
[project]
name = "demo"
[worktree]
naming = "fixed"
`,
		"bad provider": `
[project]
name = "demo"
[providers]
infra = "nomad"
`,
		"bad port range": `
[project]
name = "demo"
[ports]
range = "9000"
`,
		"inverted port range": `
[project]
name = "demo"
[ports]
range = "4000-3000"
`,
		"unknown agent runner": `
[project]
name = "demo"
[[agents]]
name = "x"
runner = "bogus"
`,
		"duplicate agent name": `
[project]
name = "demo"
[[agents]]
name = "a"
runner = "claude"
[[agents]]
name = "a"
runner = "codex"
`,
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeProject([]byte(src))
			require.Error(t, err)
		})
	}
}

func TestPortsBounds(t *testing.T) {
	lo, hi, err := PortsSection{Range: "3000-3999"}.Bounds()
	require.NoError(t, err)
	require.Equal(t, 3000, lo)
	require.Equal(t, 3999, hi)

	_, _, err = PortsSection{Range: "0-100"}.Bounds()
	require.Error(t, err, "out-of-range low bound should fail")
}

func TestLoadProjectFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ProjectFileName)
	require.NoError(t, os.WriteFile(path, []byte("[project]\nname = \"fromfile\"\n"), 0o644))

	p, err := LoadProject(path)
	require.NoError(t, err)
	require.Equal(t, "fromfile", p.Project.Name)

	_, err = LoadProject(filepath.Join(dir, "missing.toml"))
	require.Error(t, err, "missing file should fail")
}
