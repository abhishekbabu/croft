// Package command implements croft's CLI verbs, one file per verb.
package command

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/abhishekbabu/croft/internal/config"
	"github.com/spf13/cobra"
)

// NewInitCmd builds the `croft init` command, which scaffolds a croft.toml for
// the current repository.
func NewInitCmd() *cobra.Command {
	var force, yes bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a croft.toml for the current repository",
		Long: "init writes a croft.toml at the repository root. By default it\n" +
			"prompts interactively; pass --yes to accept defaults non-interactively.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return doInit(cwd, force, yes, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing croft.toml")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "accept defaults without prompting")
	return cmd
}

// doInit performs the scaffold. It is separated from the cobra command so it
// can be exercised directly in tests with explicit dir/IO.
func doInit(startDir string, force, yes bool, in io.Reader, out io.Writer) error {
	root, err := gitRepoRoot(startDir)
	if err != nil {
		return err
	}
	target := filepath.Join(root, config.ProjectFileName)
	if _, err := os.Stat(target); err == nil && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", config.ProjectFileName)
	}

	p := defaultProjectConfig(root)
	if !yes {
		gatherAnswers(&p, bufio.NewReader(in), out)
	}

	rendered := config.Scaffold(p)
	if _, err := config.DecodeProject([]byte(rendered)); err != nil {
		return fmt.Errorf("generated config is invalid: %w", err)
	}
	if err := os.WriteFile(target, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}
	fmt.Fprintf(out, "Wrote %s\n", target)
	fmt.Fprintln(out, "Review it, then run `croft new <branch>` to create your first worktree.")
	return nil
}

// defaultProjectConfig returns a sensible starting config for a repo at root.
func defaultProjectConfig(root string) config.ProjectConfig {
	p := config.ProjectConfig{
		Project: config.ProjectSection{Name: filepath.Base(root)},
		Ports:   config.PortsSection{Range: "3000-3999"},
		Agents: []config.AgentConfig{
			{Name: "claude", Runner: "claude"},
			{Name: "codex", Runner: "codex"},
		},
	}
	// applyDefaults is unexported; round-trip through Scaffold/Decode is how
	// callers normally get defaults. Here we set the few we need explicitly.
	p.Worktree.Root = "../worktrees"
	p.Worktree.Naming = p.Project.Name + ".{slug}"
	p.Providers = config.ProvidersSection{
		Multiplexer: "none", Infra: "none", Router: "none",
		Stacker: "none", Coordination: "basic",
	}
	return p
}

// gatherAnswers mutates p with interactively-collected values.
func gatherAnswers(p *config.ProjectConfig, in *bufio.Reader, out io.Writer) {
	pr := prompter{in: in, out: out}
	p.Project.Name = pr.text("Project name", p.Project.Name)
	p.Worktree.Naming = p.Project.Name + ".{slug}"
	p.Worktree.Root = pr.text("Worktree root (relative to repo)", p.Worktree.Root)
	p.Worktree.DevCommand = pr.text("Dev server command ({port} is substituted)", p.Worktree.DevCommand)
	p.Providers.Multiplexer = pr.choice("Multiplexer", config.Multiplexers, p.Providers.Multiplexer)
	p.Providers.Infra = pr.choice("Infra", config.InfraProviders, p.Providers.Infra)
	p.Providers.Router = pr.choice("Router", config.Routers, p.Providers.Router)
	p.Providers.Stacker = pr.choice("Stacker", config.Stackers, p.Providers.Stacker)
}

// gitRepoRoot resolves the top level of the git repo containing dir.
func gitRepoRoot(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository (run `git init` first)")
	}
	return strings.TrimSpace(string(out)), nil
}

// prompter reads simple line-based answers from a reader.
type prompter struct {
	in  *bufio.Reader
	out io.Writer
}

// text asks for a free-text value, returning def when the answer is blank.
func (p prompter) text(label, def string) string {
	if def != "" {
		fmt.Fprintf(p.out, "%s [%s]: ", label, def)
	} else {
		fmt.Fprintf(p.out, "%s: ", label)
	}
	line, _ := p.in.ReadString('\n')
	if line = strings.TrimSpace(line); line == "" {
		return def
	}
	return line
}

// choice asks for a value constrained to options, re-prompting on a bad answer.
func (p prompter) choice(label string, options []string, def string) string {
	prompt := fmt.Sprintf("%s (%s)", label, strings.Join(options, "/"))
	for {
		v := p.text(prompt, def)
		for _, o := range options {
			if v == o {
				return v
			}
		}
		fmt.Fprintf(p.out, "  %q is not one of %v\n", v, options)
	}
}
