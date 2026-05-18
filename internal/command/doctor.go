package command

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/abhishekbabu/croft/internal/provider"
	"github.com/abhishekbabu/croft/internal/sh"
	"github.com/abhishekbabu/croft/internal/worktree"
	"github.com/spf13/cobra"
)

// finding is one issue doctor detected, optionally with an automatic fix.
type finding struct {
	desc string
	fix  func() error // nil when there is no safe automatic fix
}

// NewDoctorCmd builds the `croft doctor` command, which reconciles croft state
// against reality.
func NewDoctorCmd() *cobra.Command {
	var fix bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Detect and reconcile orphaned worktrees, leaked resources, and stale state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(cwd)
			if err != nil {
				return err
			}
			return doDoctor(ctx, fix, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "apply fixes for the issues found")
	return cmd
}

// doDoctor runs every reconcile check and reports (or, with fix, repairs).
func doDoctor(ctx *appContext, fix bool, out io.Writer) error {
	// The checks are independent and each shells out (git, docker); run them
	// concurrently and reassemble findings in declaration order.
	checks := []func(*appContext) []finding{
		checkStaleRegistry,
		checkGitWorktrees,
		checkOrphanDirs,
		checkLeakedContainers,
	}
	results := make([][]finding, len(checks))
	var wg sync.WaitGroup
	for i, check := range checks {
		wg.Add(1)
		go func(i int, check func(*appContext) []finding) {
			defer wg.Done()
			results[i] = check(ctx)
		}(i, check)
	}
	wg.Wait()

	var findings []finding
	for _, r := range results {
		findings = append(findings, r...)
	}

	if len(findings) == 0 {
		fmt.Fprintln(out, "All clear — no issues found.")
		return nil
	}

	fmt.Fprintf(out, "Found %d issue(s):\n", len(findings))
	var failed int
	for _, f := range findings {
		switch {
		case !fix:
			fmt.Fprintf(out, "  - %s\n", f.desc)
		case f.fix == nil:
			fmt.Fprintf(out, "  - %s (no automatic fix)\n", f.desc)
		default:
			if err := f.fix(); err != nil {
				fmt.Fprintf(out, "  x %s: fix failed: %v\n", f.desc, err)
				failed++
			} else {
				fmt.Fprintf(out, "  fixed: %s\n", f.desc)
			}
		}
	}

	if !fix {
		fmt.Fprintln(out, "Re-run with --fix to apply fixes.")
		return nil
	}
	if failed > 0 {
		return fmt.Errorf("%d fix(es) failed", failed)
	}
	return nil
}

// checkStaleRegistry finds registry entries whose worktree directory is gone.
func checkStaleRegistry(ctx *appContext) []finding {
	reg, err := ctx.Store.Load()
	if err != nil {
		return nil
	}
	var fs []finding
	for slug, wt := range reg.Worktrees {
		if dirExists(wt.Path) {
			continue
		}
		slug, wt := slug, wt
		fs = append(fs, finding{
			desc: fmt.Sprintf("registry entry %q points at a missing directory (%s)", slug, wt.Path),
			fix: func() error {
				pw := ctx.providerWorktree(wt)
				_ = ctx.Providers.Infra.Down(pw)
				_ = ctx.Providers.Router.Release(pw)
				_ = ctx.Providers.Multiplexer.Kill(provider.ProjectName(pw))
				return ctx.Store.Delete(slug)
			},
		})
	}
	return fs
}

// checkGitWorktrees finds git worktree administrative entries whose directory
// no longer exists.
func checkGitWorktrees(ctx *appContext) []finding {
	list, err := ctx.Manager.List()
	if err != nil {
		return nil
	}
	var stale []string
	for _, w := range list {
		if w.Bare {
			continue
		}
		if !dirExists(w.Path) {
			stale = append(stale, w.Path)
		}
	}
	if len(stale) == 0 {
		return nil
	}
	return []finding{{
		desc: fmt.Sprintf("git has %d stale worktree entr(ies): %s", len(stale), strings.Join(stale, ", ")),
		fix:  ctx.Manager.Prune,
	}}
}

// checkOrphanDirs finds directories under the worktree root that match the
// naming pattern but have no registry entry.
func checkOrphanDirs(ctx *appContext) []finding {
	entries, err := os.ReadDir(ctx.WorktreeRoot)
	if err != nil {
		return nil // the worktree root may not exist yet
	}
	reg, err := ctx.Store.Load()
	if err != nil {
		return nil
	}
	var fs []finding
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		slug, ok := worktree.SlugFromDir(e.Name(), ctx.Config.Worktree.Naming)
		if !ok {
			continue
		}
		if _, registered := reg.Worktrees[slug]; registered {
			continue
		}
		path := filepath.Join(ctx.WorktreeRoot, e.Name())
		fs = append(fs, finding{
			desc: fmt.Sprintf("orphan worktree directory %q (matches naming pattern, not in registry)", path),
			fix: func() error {
				if err := ctx.Manager.Remove(path, true); err != nil {
					if rmErr := os.RemoveAll(path); rmErr != nil {
						return rmErr
					}
				}
				return ctx.Manager.Prune()
			},
		})
	}
	return fs
}

// checkLeakedContainers finds docker compose stacks named for this project
// that have no matching worktree.
func checkLeakedContainers(ctx *appContext) []finding {
	if !sh.Look("docker") {
		return nil
	}
	out, err := sh.Capture("docker", "", nil, "compose", "ls", "--all", "--format", "json")
	if err != nil {
		return nil
	}
	var projects []struct {
		Name string `json:"Name"`
	}
	if json.Unmarshal([]byte(out), &projects) != nil {
		return nil
	}
	reg, err := ctx.Store.Load()
	if err != nil {
		return nil
	}
	prefix := ctx.Config.Project.Name + "-"
	var fs []finding
	for _, p := range projects {
		if !strings.HasPrefix(p.Name, prefix) {
			continue
		}
		slug := strings.TrimPrefix(p.Name, prefix)
		if _, ok := reg.Worktrees[slug]; ok {
			continue
		}
		name := p.Name
		fs = append(fs, finding{
			desc: fmt.Sprintf("leaked container stack %q (no matching worktree)", name),
			fix: func() error {
				_, err := sh.Capture("docker", "", nil, "compose", "-p", name, "down", "-v")
				return err
			},
		})
	}
	return fs
}
