package command

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/abhishekbabu/croft/internal/config"
	"github.com/abhishekbabu/croft/internal/provider"
	"github.com/abhishekbabu/croft/internal/state"
	"github.com/abhishekbabu/croft/internal/worktree"
)

// appContext bundles everything a command needs: the resolved repository,
// parsed configuration, the state store, a git-worktree manager, and the
// configured provider set.
type appContext struct {
	RepoRoot     string
	WorktreeRoot string // absolute path where checkouts land
	Config       config.ProjectConfig
	Machine      config.MachineConfig
	Store        *state.Store
	Manager      *worktree.Manager
	Providers    provider.Set
}

// loadContext resolves the croft context for the repository containing
// startDir.
func loadContext(startDir string) (*appContext, error) {
	root, err := gitRepoRoot(startDir)
	if err != nil {
		return nil, err
	}
	cfgPath := filepath.Join(root, config.ProjectFileName)
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no %s found at %s (run `croft init`)", config.ProjectFileName, root)
	}
	cfg, err := config.LoadProject(cfgPath)
	if err != nil {
		return nil, err
	}
	machine, err := config.LoadMachine()
	if err != nil {
		return nil, err
	}
	store, err := state.Open(cfg.Project.Name)
	if err != nil {
		return nil, err
	}
	providers, err := provider.New(cfg.Providers, machine)
	if err != nil {
		return nil, err
	}
	return &appContext{
		RepoRoot:     root,
		WorktreeRoot: filepath.Clean(filepath.Join(root, cfg.Worktree.Root)),
		Config:       cfg,
		Machine:      machine,
		Store:        store,
		Manager:      worktree.NewManager(root),
		Providers:    providers,
	}, nil
}

// providerWorktree adapts a registry record into the value providers operate
// on.
func (c *appContext) providerWorktree(wt state.Worktree) provider.Worktree {
	return provider.Worktree{
		Project: c.Config.Project.Name,
		Slug:    wt.Slug,
		Path:    wt.Path,
		Ports:   wt.Ports,
	}
}

// devCommand returns the worktree's dev server command with the {port}
// placeholder substituted, or an empty string when none is configured.
func (c *appContext) devCommand(wt state.Worktree) string {
	cmd := c.Config.Worktree.DevCommand
	if cmd == "" {
		return ""
	}
	return strings.ReplaceAll(cmd, "{port}", c.primaryPort(wt))
}

// primaryPort returns the port of the first configured service, as a string.
func (c *appContext) primaryPort(wt state.Worktree) string {
	for _, svc := range c.Config.Ports.Services {
		if p, ok := wt.Ports[svc]; ok {
			return strconv.Itoa(p)
		}
	}
	return ""
}

// takenPorts returns the union of every port already allocated in the registry.
func takenPorts(r state.Registry) map[int]bool {
	taken := map[int]bool{}
	for _, wt := range r.Worktrees {
		for _, p := range wt.Ports {
			taken[p] = true
		}
	}
	return taken
}

// dirExists reports whether path is an existing directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// formatPorts renders a port map as a stable, space-separated string.
func formatPorts(ports map[string]int) string {
	if len(ports) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(ports))
	for k := range ports {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s=%d", k, ports[k])
	}
	return strings.Join(parts, " ")
}
