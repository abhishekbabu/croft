// Package config loads croft's two configuration layers: the committed,
// repo-root project config (croft.toml) and the per-machine config
// (~/.config/croft/config.toml). See PLAN.md §6.2.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// ProjectFileName is the committed, repo-root project config file.
const ProjectFileName = "croft.toml"

// ProjectConfig is the parsed, defaulted contents of croft.toml. It is shared
// by the whole team and committed to the repo.
type ProjectConfig struct {
	Project   ProjectSection   `toml:"project"`
	Worktree  WorktreeSection  `toml:"worktree"`
	Ports     PortsSection     `toml:"ports"`
	Providers ProvidersSection `toml:"providers"`
	Agents    []AgentConfig    `toml:"agents"`
	Hooks     HooksSection     `toml:"hooks"`
}

// ProjectSection identifies the project.
type ProjectSection struct {
	Name string `toml:"name"`
}

// WorktreeSection describes where and how worktree checkouts are created.
type WorktreeSection struct {
	Root       string   `toml:"root"`        // where checkouts land, relative to repo root
	Naming     string   `toml:"naming"`      // directory naming pattern, must contain {slug}
	DevCommand string   `toml:"dev_command"` // dev server command; {port} is substituted
	CopyFiles  []string `toml:"copy_files"`  // untracked files seeded into each worktree
}

// PortsSection describes the per-worktree port allocation.
type PortsSection struct {
	Range    string   `toml:"range"`    // inclusive port range, e.g. "3000-3999"
	Services []string `toml:"services"` // services that each get a unique port
}

// ProvidersSection selects the backend implementation for each provider slot.
type ProvidersSection struct {
	Multiplexer  string `toml:"multiplexer"`
	Infra        string `toml:"infra"`
	Router       string `toml:"router"`
	Stacker      string `toml:"stacker"`
	Coordination string `toml:"coordination"`
}

// AgentConfig names an agent and the runner that backs it. Command is the
// argv template used only by the "exec" runner ({dir} and {prompt} are
// substituted at launch).
type AgentConfig struct {
	Name    string   `toml:"name"`
	Runner  string   `toml:"runner"`
	Command []string `toml:"command"`
}

// HooksSection holds shell commands run around worktree lifecycle events.
type HooksSection struct {
	PostCreate []string `toml:"post_create"`
	PreRemove  []string `toml:"pre_remove"`
}

// LoadProject reads, decodes, defaults, and validates the project config at
// path.
func LoadProject(path string) (ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ProjectConfig{}, fmt.Errorf("read %s: %w", path, err)
	}
	return DecodeProject(data)
}

// DecodeProject decodes a project config from raw TOML bytes, applies defaults,
// and validates the result.
func DecodeProject(data []byte) (ProjectConfig, error) {
	var p ProjectConfig
	if _, err := toml.Decode(string(data), &p); err != nil {
		return ProjectConfig{}, fmt.Errorf("parse %s: %w", ProjectFileName, err)
	}
	p.applyDefaults()
	if err := p.Validate(); err != nil {
		return ProjectConfig{}, err
	}
	return p, nil
}

// applyDefaults fills unset fields with their conventional defaults. It is
// safe to call more than once.
func (p *ProjectConfig) applyDefaults() {
	if p.Worktree.Root == "" {
		p.Worktree.Root = "../worktrees"
	}
	if p.Worktree.Naming == "" {
		if p.Project.Name != "" {
			p.Worktree.Naming = p.Project.Name + ".{slug}"
		} else {
			p.Worktree.Naming = "{slug}"
		}
	}
	if p.Providers.Multiplexer == "" {
		p.Providers.Multiplexer = "none"
	}
	if p.Providers.Infra == "" {
		p.Providers.Infra = "none"
	}
	if p.Providers.Router == "" {
		p.Providers.Router = "none"
	}
	if p.Providers.Stacker == "" {
		p.Providers.Stacker = "none"
	}
	if p.Providers.Coordination == "" {
		p.Providers.Coordination = "basic"
	}
}

// Validate reports the first configuration error it finds.
func (p *ProjectConfig) Validate() error {
	if strings.TrimSpace(p.Project.Name) == "" {
		return fmt.Errorf("project.name: required")
	}
	if !strings.Contains(p.Worktree.Naming, "{slug}") {
		return fmt.Errorf("worktree.naming: %q must contain the {slug} placeholder", p.Worktree.Naming)
	}
	if p.Ports.Range != "" {
		if _, _, err := p.Ports.Bounds(); err != nil {
			return err
		}
	}
	for field, check := range map[string][2]string{
		"providers.multiplexer":  {p.Providers.Multiplexer, "multiplexer"},
		"providers.infra":        {p.Providers.Infra, "infra"},
		"providers.router":       {p.Providers.Router, "router"},
		"providers.stacker":      {p.Providers.Stacker, "stacker"},
		"providers.coordination": {p.Providers.Coordination, "coordination"},
	} {
		if err := validateProvider(field, check[0], check[1]); err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	for i, a := range p.Agents {
		if strings.TrimSpace(a.Name) == "" {
			return fmt.Errorf("agents[%d].name: required", i)
		}
		if seen[a.Name] {
			return fmt.Errorf("agents[%d].name: %q is duplicated", i, a.Name)
		}
		seen[a.Name] = true
		if err := validateEnum(fmt.Sprintf("agents[%d].runner", i), a.Runner, AgentRunners); err != nil {
			return err
		}
		if a.Runner == "exec" && len(a.Command) == 0 {
			return fmt.Errorf("agents[%d].command: required when runner is \"exec\"", i)
		}
	}
	return nil
}

// validateProvider dispatches enum validation to the right allowed-value set.
func validateProvider(field, value, slot string) error {
	switch slot {
	case "multiplexer":
		return validateEnum(field, value, Multiplexers)
	case "infra":
		return validateEnum(field, value, InfraProviders)
	case "router":
		return validateEnum(field, value, Routers)
	case "stacker":
		return validateEnum(field, value, Stackers)
	case "coordination":
		return validateEnum(field, value, Coordinations)
	default:
		return fmt.Errorf("unknown provider slot %q", slot)
	}
}

// Bounds parses the port range into inclusive low/high bounds.
func (s PortsSection) Bounds() (low, high int, err error) {
	lo, hi, ok := strings.Cut(s.Range, "-")
	if !ok {
		return 0, 0, fmt.Errorf("ports.range: %q must be of the form \"LOW-HIGH\"", s.Range)
	}
	low, err = strconv.Atoi(strings.TrimSpace(lo))
	if err != nil {
		return 0, 0, fmt.Errorf("ports.range: invalid low bound %q", lo)
	}
	high, err = strconv.Atoi(strings.TrimSpace(hi))
	if err != nil {
		return 0, 0, fmt.Errorf("ports.range: invalid high bound %q", hi)
	}
	if low < 1 || high > 65535 {
		return 0, 0, fmt.Errorf("ports.range: %q out of the valid 1-65535 range", s.Range)
	}
	if low >= high {
		return 0, 0, fmt.Errorf("ports.range: low bound %d must be below high bound %d", low, high)
	}
	return low, high, nil
}
