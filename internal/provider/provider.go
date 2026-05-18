// Package provider defines croft's five swappable backend interfaces —
// multiplexer, infra, router, stacker, coordination — and their built-in
// implementations. Every interface has a no-op default so a minimal setup
// needs no external tools (PLAN.md §6.3).
package provider

import (
	"fmt"
	"regexp"
	"strings"
)

// Worktree is the subset of a worktree's state that providers operate on.
type Worktree struct {
	Project string
	Slug    string
	Path    string
	Ports   map[string]int
}

// ProjectName returns the per-worktree namespace shared by the container stack
// and the terminal session — "<project>-<slug>", so neither collides across
// worktrees or across projects.
func ProjectName(wt Worktree) string {
	project := wt.Project
	if project == "" {
		project = "croft"
	}
	return project + "-" + wt.Slug
}

var envUnsafe = regexp.MustCompile(`[^A-Z0-9]+`)

// Env returns the environment variables croft injects into a worktree's
// session and container stack: identity, paths, the compose project name, and
// one <SERVICE>_PORT per allocated port.
func Env(wt Worktree) map[string]string {
	env := map[string]string{
		"CROFT_SLUG":           wt.Slug,
		"CROFT_WORKTREE":       wt.Path,
		"COMPOSE_PROJECT_NAME": ProjectName(wt),
	}
	for svc, port := range wt.Ports {
		key := envUnsafe.ReplaceAllString(strings.ToUpper(svc), "_")
		env[key+"_PORT"] = fmt.Sprintf("%d", port)
	}
	return env
}

// envSlice renders an environment map as KEY=VALUE pairs.
func envSlice(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}
