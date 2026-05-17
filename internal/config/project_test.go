package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeProjectAppliesDefaults(t *testing.T) {
	p, err := DecodeProject([]byte(`
[project]
name = "demo"
`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if p.Worktree.Root != "../worktrees" {
		t.Errorf("worktree.root default = %q, want ../worktrees", p.Worktree.Root)
	}
	if p.Worktree.Naming != "demo.{slug}" {
		t.Errorf("worktree.naming default = %q, want demo.{slug}", p.Worktree.Naming)
	}
	if p.Providers.Multiplexer != "none" {
		t.Errorf("providers.multiplexer default = %q, want none", p.Providers.Multiplexer)
	}
	if p.Providers.Coordination != "basic" {
		t.Errorf("providers.coordination default = %q, want basic", p.Providers.Coordination)
	}
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
			if _, err := DecodeProject([]byte(src)); err == nil {
				t.Fatalf("expected validation error, got nil")
			}
		})
	}
}

func TestPortsBounds(t *testing.T) {
	lo, hi, err := PortsSection{Range: "3000-3999"}.Bounds()
	if err != nil {
		t.Fatalf("Bounds: %v", err)
	}
	if lo != 3000 || hi != 3999 {
		t.Errorf("Bounds = %d,%d, want 3000,3999", lo, hi)
	}
	if _, _, err := (PortsSection{Range: "0-100"}).Bounds(); err == nil {
		t.Error("expected error for out-of-range low bound")
	}
}

func TestLoadProjectFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ProjectFileName)
	if err := os.WriteFile(path, []byte("[project]\nname = \"fromfile\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := LoadProject(path)
	if err != nil {
		t.Fatalf("LoadProject: %v", err)
	}
	if p.Project.Name != "fromfile" {
		t.Errorf("name = %q, want fromfile", p.Project.Name)
	}
	if _, err := LoadProject(filepath.Join(dir, "missing.toml")); err == nil {
		t.Error("expected error for missing file")
	}
}
