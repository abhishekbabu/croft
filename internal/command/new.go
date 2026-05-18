package command

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/abhishekbabu/croft/internal/provider"
	"github.com/abhishekbabu/croft/internal/state"
	"github.com/abhishekbabu/croft/internal/worktree"
	"github.com/spf13/cobra"
)

// NewNewCmd builds the `croft new` command, which creates a fully isolated
// environment for a branch: worktree, port set, container stack, and session.
func NewNewCmd() *cobra.Command {
	var from, agentName string
	cmd := &cobra.Command{
		Use:   "new <branch>",
		Short: "Create an isolated environment for a branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCwd()
			if err != nil {
				return err
			}
			return doNew(ctx, args[0], from, agentName, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "start point for a new branch (default: current HEAD)")
	cmd.Flags().StringVar(&agentName, "agent", "", "launch a configured agent into the worktree")
	return cmd
}

// doNew creates the worktree for branch and reconciles its environment. It is
// idempotent: re-running it on an existing worktree re-converges the container
// stack and session rather than failing. When agentName is set, the agent is
// launched into the worktree's session.
func doNew(ctx *appContext, branch, from, agentName string, out io.Writer) error {
	slug := worktree.Slugify(branch)
	if slug == "" {
		return fmt.Errorf("branch %q produces an empty slug", branch)
	}

	reg, err := ctx.Store.Load()
	if err != nil {
		return err
	}
	rec, registered := reg.Worktrees[slug]

	// Create the worktree itself unless it already exists on disk.
	if !registered || !dirExists(rec.Path) {
		rec, err = createWorktree(ctx, reg, slug, branch, from, rec)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Created worktree %q (branch %s)\n", slug, branch)
		seedCopyFiles(ctx.RepoRoot, rec.Path, ctx.Config.Worktree.CopyFiles, out)
	} else {
		fmt.Fprintf(out, "Reconciling worktree %q\n", slug)
	}
	fmt.Fprintf(out, "  path:  %s\n", rec.Path)
	if len(rec.Ports) > 0 {
		fmt.Fprintf(out, "  ports: %s\n", formatPorts(rec.Ports))
	}

	// Reconcile the isolated environment. These steps are all idempotent.
	pw := ctx.providerWorktree(rec)
	env := provider.Env(pw)
	if err := ctx.Providers.Multiplexer.CreateSession(provider.ProjectName(pw), rec.Path, env); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	if err := ctx.Providers.Infra.Up(pw); err != nil {
		return fmt.Errorf("bring infra up: %w", err)
	}
	url, err := ctx.Providers.Router.Register(pw)
	if err != nil {
		return fmt.Errorf("register route: %w", err)
	}
	if url != rec.URL {
		rec.URL = url
		if err := ctx.Store.Put(rec); err != nil {
			return err
		}
	}
	if url != "" {
		fmt.Fprintf(out, "  url:   %s\n", url)
	}
	if err := runHooks("post_create", ctx.Config.Hooks.PostCreate, rec.Path, env, out); err != nil {
		return err
	}
	if err := startDevServer(ctx, rec, env, out); err != nil {
		return err
	}

	if agentName != "" {
		if err := launchAgent(ctx, rec, agentName, env, out); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Worktree %q is ready.\n", slug)
	return nil
}

// startDevServer runs the configured dev server in the worktree's session,
// idempotently. With no multiplexer to host it, it prints the command instead.
func startDevServer(ctx *appContext, rec state.Worktree, env map[string]string, out io.Writer) error {
	cmd := ctx.devCommand(rec)
	if cmd == "" {
		return nil
	}
	mux := ctx.Providers.Multiplexer
	if !mux.Managed() {
		fmt.Fprintf(out, "  dev server: no multiplexer configured — start it yourself:\n    %s\n", cmd)
		return nil
	}
	session := provider.ProjectName(ctx.providerWorktree(rec))
	if mux.HasWindow(session, windowDev) {
		fmt.Fprintln(out, "  dev server: already running")
		return nil
	}
	if err := mux.RunWindow(session, windowDev, rec.Path, env, []string{"sh", "-c", cmd}); err != nil {
		return fmt.Errorf("start dev server: %w", err)
	}
	fmt.Fprintf(out, "  dev server: started (%s)\n", cmd)
	return nil
}

// createWorktree adds the git worktree, allocates ports, and records the
// registry entry transactionally (a failed state write rolls the worktree
// back).
func createWorktree(ctx *appContext, reg state.Registry, slug, branch, from string, prev state.Worktree) (state.Worktree, error) {
	id := worktree.Resolve(slug, ctx.WorktreeRoot, ctx.Config.Worktree.Naming)
	// #nosec G301 -- the worktree root holds ordinary source checkouts.
	if err := os.MkdirAll(ctx.WorktreeRoot, 0o755); err != nil {
		return state.Worktree{}, fmt.Errorf("create worktree root: %w", err)
	}

	// Reuse a prior allocation when reconciling a stale entry; otherwise
	// allocate a fresh port set.
	ports := prev.Ports
	if len(ports) == 0 && ctx.Config.Ports.Range != "" && len(ctx.Config.Ports.Services) > 0 {
		low, high, err := ctx.Config.Ports.Bounds()
		if err != nil {
			return state.Worktree{}, err
		}
		ports, err = worktree.AllocatePorts(low, high, ctx.Config.Ports.Services, takenPorts(reg))
		if err != nil {
			return state.Worktree{}, err
		}
	}

	if err := ctx.Manager.Add(id.Path, branch, from); err != nil {
		return state.Worktree{}, err
	}

	rec := state.Worktree{
		Slug:    slug,
		Branch:  branch,
		Path:    id.Path,
		Ports:   ports,
		Status:  prev.Status,
		Created: prev.Created,
	}
	if rec.Created.IsZero() {
		rec.Created = time.Now().UTC()
	}
	if err := ctx.Store.Put(rec); err != nil {
		_ = ctx.Manager.Remove(id.Path, true)
		return state.Worktree{}, fmt.Errorf("record worktree (rolled back): %w", err)
	}
	return rec, nil
}

// seedCopyFiles copies each configured copy_files entry from the repo root
// into a freshly created worktree, at the same relative path. The entries are
// untracked developer files (an .env.local, a local override), so a missing
// source is skipped silently; a genuine copy failure is reported but never
// fails worktree creation.
func seedCopyFiles(repoRoot, worktreePath string, files []string, out io.Writer) {
	for _, rel := range files {
		src := filepath.Join(repoRoot, rel)
		info, err := os.Stat(src)
		if err != nil {
			continue // not present — expected for optional dev files
		}
		if err := copyFile(src, filepath.Join(worktreePath, rel), info.Mode()); err != nil {
			fmt.Fprintf(out, "  warning: could not seed %s: %v\n", rel, err)
			continue
		}
		fmt.Fprintf(out, "  seeded: %s\n", rel)
	}
}

// copyFile copies src to dst, creating parent directories, keeping src's
// permissions so a copied file stays exactly as the developer left it.
func copyFile(src, dst string, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	// #nosec G306 -- a copied file keeps its source's permissions.
	return os.WriteFile(dst, data, perm)
}
