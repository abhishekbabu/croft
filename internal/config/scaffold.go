package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Scaffold renders a commented croft.toml from a ProjectConfig. The output is
// human-friendly (comments, stable section order) rather than a plain struct
// marshal, since it is meant to be committed and hand-edited.
func Scaffold(p ProjectConfig) string {
	var b strings.Builder
	w := func(format string, args ...any) {
		fmt.Fprintf(&b, format, args...)
	}

	w("# croft project configuration — committed, shared by the whole team.\n")
	w("# Per-machine settings live in ~/.config/croft/config.toml instead.\n\n")

	w("[project]\n")
	w("name = %s\n\n", tomlString(p.Project.Name))

	w("[worktree]\n")
	w("# Where checkouts land, relative to the repo root.\n")
	w("root = %s\n", tomlString(p.Worktree.Root))
	w("# Directory naming pattern; {slug} is the worktree's stable identity.\n")
	w("naming = %s\n", tomlString(p.Worktree.Naming))
	w("# Dev server command; {port} is substituted per worktree.\n")
	w("dev_command = %s\n", tomlString(p.Worktree.DevCommand))
	w("# Untracked files seeded into each new worktree.\n")
	w("copy_files = %s\n\n", tomlStringArray(p.Worktree.CopyFiles))

	w("[ports]\n")
	w("# Inclusive port range each worktree draws a unique set from.\n")
	w("range = %s\n", tomlString(p.Ports.Range))
	w("services = %s\n\n", tomlStringArray(p.Ports.Services))

	w("[providers]\n")
	w("multiplexer  = %s  # %s\n", tomlString(p.Providers.Multiplexer), strings.Join(Multiplexers, " | "))
	w("infra        = %s  # %s\n", tomlString(p.Providers.Infra), strings.Join(InfraProviders, " | "))
	w("router       = %s  # %s\n", tomlString(p.Providers.Router), strings.Join(Routers, " | "))
	w("stacker      = %s  # %s\n", tomlString(p.Providers.Stacker), strings.Join(Stackers, " | "))
	w("coordination = %s  # %s\n\n", tomlString(p.Providers.Coordination), strings.Join(Coordinations, " | "))

	for _, a := range p.Agents {
		w("[[agents]]\n")
		w("name   = %s\n", tomlString(a.Name))
		w("runner = %s\n\n", tomlString(a.Runner))
	}

	w("[hooks]\n")
	w("# Shell commands run after a worktree is created / before it is removed.\n")
	w("post_create = %s\n", tomlStringArray(p.Hooks.PostCreate))
	w("pre_remove  = %s\n", tomlStringArray(p.Hooks.PreRemove))

	return b.String()
}

// tomlString quotes a string as a TOML basic string.
func tomlString(s string) string {
	return strconv.Quote(s)
}

// tomlStringArray renders a string slice as a single-line TOML array.
func tomlStringArray(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = tomlString(s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
