# Changelog

All notable changes to croft are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/), and the project adheres to
[Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.0]

First release — the Phase 1 isolation layer.

### Added

- `croft init` — scaffold a `croft.toml` for a repository.
- `croft new` — create a fully isolated environment for a branch: git
  worktree, a unique port set, a container stack, a terminal session, a route,
  and optionally an AI coding agent (`--agent`). Idempotent and transactional.
- `croft ls` / `croft status` — list worktrees and inspect one.
- `croft sync` — rebase branch stacks against the trunk: refuses mid-rebase,
  auto-stashes, and applies a stack-resolved teardown gate.
- `croft rm` — idempotent teardown of a worktree and its environment.
- `croft doctor` — reconcile croft state against reality (orphans, leaks,
  stale entries), with `--fix`.
- `croft spawn` / `croft fleet` — spawn and coordinate named agents across
  worktrees.
- Two-layer TOML configuration (project + per-machine) and JSON runtime state
  under the XDG data directory.
- Five swappable provider interfaces — multiplexer (`tmux`/`cmux`), infra
  (`docker-compose`), router (`portless`), stacker (`graphite`), coordination
  (`basic`) — each with a no-op default.
- Agent runners for Claude, Codex, and any generic CLI agent.

[Unreleased]: https://github.com/abhishekbabu/croft/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/abhishekbabu/croft/releases/tag/v0.1.0
