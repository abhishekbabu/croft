package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/abhishekbabu/croft/internal/env"
)

// MachineFileName is the per-machine config file name.
const MachineFileName = "config.toml"

// MachineConfig is the per-machine, gitignored config — everything that would
// otherwise be hardcoded to one developer's setup (binary paths, AWS profile,
// personal defaults). See PLAN.md §6.2.
type MachineConfig struct {
	Bins     map[string]string `toml:"bins"`     // tool name -> absolute binary path
	Defaults DefaultsSection   `toml:"defaults"` // personal defaults
	AWS      AWSSection        `toml:"aws"`      // AWS SSO pre-check settings
}

// DefaultsSection holds the developer's personal defaults.
type DefaultsSection struct {
	Agent  string `toml:"agent"`
	Effort string `toml:"effort"`
}

// AWSSection holds machine-specific AWS settings.
type AWSSection struct {
	SSOProfile string `toml:"sso_profile"`
}

// MachineConfigPath returns the path to the per-machine config file, honoring
// XDG_CONFIG_HOME and falling back to ~/.config.
func MachineConfigPath() (string, error) {
	dir, err := env.ConfigHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "croft", MachineFileName), nil
}

// LoadMachine reads the per-machine config. The file is optional: if it does
// not exist, a zero-value config is returned with no error.
func LoadMachine() (MachineConfig, error) {
	path, err := MachineConfigPath()
	if err != nil {
		return MachineConfig{}, err
	}
	return LoadMachineFrom(path)
}

// LoadMachineFrom reads the per-machine config from an explicit path. A missing
// file is not an error.
func LoadMachineFrom(path string) (MachineConfig, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return MachineConfig{}, nil
	}
	if err != nil {
		return MachineConfig{}, fmt.Errorf("read %s: %w", path, err)
	}
	var m MachineConfig
	if _, err := toml.Decode(string(data), &m); err != nil {
		return MachineConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return m, nil
}

// Bin returns the configured path for a tool, or the tool name itself when no
// override is set (so it is resolved from PATH).
func (m MachineConfig) Bin(tool string) string {
	if p, ok := m.Bins[tool]; ok && p != "" {
		return p
	}
	return tool
}

// DefaultAgent resolves the effective default agent: the machine override if
// set, otherwise the first agent declared in the project config, otherwise an
// empty string.
func (m MachineConfig) DefaultAgent(p ProjectConfig) string {
	if m.Defaults.Agent != "" {
		return m.Defaults.Agent
	}
	if len(p.Agents) > 0 {
		return p.Agents[0].Name
	}
	return ""
}
