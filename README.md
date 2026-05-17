# croft

**Worktrees that actually work.**

croft gives every git branch its own fully isolated, working development
environment — its own checkout, container stack, ports, route, dev server, and
(optionally) an AI coding agent — created and torn down in seconds.

## Why

Plenty of tools spawn AI agents into git worktrees. Nearly all of them stop at
creating the worktree and punt on **runtime isolation**: worktrees collide on
ports, share one `docker compose` namespace, and write to the same dev
database. croft is the isolation layer underneath — it makes a worktree a
*real*, self-contained environment that a developer or an agent can actually
run.

## Features

- **One isolated environment per branch** — dedicated checkout, container
  stack, port set, route, and dev server.
- **Stable, path-derived identity** — navigating a branch stack never changes a
  worktree's resources or state.
- **Pluggable providers** — multiplexer, infra, router, and stacker backends
  are swappable; every one has a no-op default, so plain git + Docker is enough.
- **Agent-ready** — launch Claude, Codex, or any CLI agent into a worktree, and
  coordinate several of them.
- **Idempotent and transactional** — every create and teardown is safe to
  re-run and cleans up after partial failures.

## Installation

```sh
go install github.com/abhishekbabu/croft/cmd/croft@latest
```

Or build from source:

```sh
git clone https://github.com/abhishekbabu/croft
cd croft
go build ./...
```

## Quick start

```sh
# Scaffold a croft.toml for the current repository
croft init

# Create an isolated environment for a branch
croft new my-feature

# See every worktree and its health
croft ls
```

## Contributing

Issues and pull requests are welcome. Run `go build ./...` and `go test ./...`
before opening a PR; CI runs the same checks plus `golangci-lint`.

## License

[Apache-2.0](LICENSE)
