# croft documentation

croft gives every git branch its own fully isolated, working development
environment — checkout, container stack, ports, route, dev server, and
optionally an AI coding agent.

## Contents

| Document | What it covers |
|----------|----------------|
| [Getting started](getting-started.md) | Install croft and create your first isolated worktree. |
| [Command reference](commands.md) | Every command, its flags, and exact behavior. |
| [Configuration](configuration.md) | `croft.toml`, the per-machine config, and runtime state. |
| [Providers](providers.md) | The five swappable backends and their `none` defaults. |
| [Agents](agents.md) | Agent runners (Claude, Codex, generic) and multi-agent coordination. |
| [Migrating from an existing setup](cutover.md) | Adopting croft incrementally, alongside existing tooling. |

## Command quick reference

| Command | Purpose |
|---------|---------|
| `croft init` | Scaffold a `croft.toml` for the current repository. |
| `croft new <branch>` | Create a fully isolated environment for a branch. |
| `croft ls` | List croft-managed worktrees and their status. |
| `croft status <branch>` | Show detail for one worktree. |
| `croft sync [branch]` | Rebase branch stacks against the trunk. |
| `croft rm <branch>` | Tear a worktree and its environment down. |
| `croft doctor` | Detect and reconcile orphans, leaks, and stale state. |
| `croft spawn <name> --agent <a>` | Spawn a named coordinated agent. |
| `croft fleet status` / `croft fleet msg` | Cross-worktree peer status and dispatch. |

Run `croft <command> --help` for flags. See the [command reference](commands.md)
for full detail.

## Concepts in one minute

- **Worktree** — a git worktree croft manages: its own checkout, port set,
  container stack, terminal session, and route.
- **Slug** — a worktree's stable identity, derived from its directory name. It
  never changes when git HEAD moves, so navigating a branch stack never changes
  a worktree's resources or state keys.
- **Provider** — a swappable backend for one concern (multiplexer, infra,
  router, stacker, coordination). Every provider has a no-op default, so a repo
  with just git and Docker works with zero extra tools.
- **Registry** — croft's JSON record of every worktree, under the XDG data
  directory. No daemon, no database.
