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
- **Pluggable providers** — multiplexer, infra, router, stacker, and
  coordination backends are swappable; every one has a no-op default, so plain
  git + Docker is enough.
- **Agent-ready** — launch Claude, Codex, or any CLI agent into a worktree, and
  coordinate several of them.
- **Idempotent and transactional** — every create and teardown is safe to
  re-run and cleans up after partial failures.

## Installation

With [Homebrew](https://brew.sh) (macOS and Linux):

```sh
brew install abhishekbabu/tap/croft
```

Install script (macOS and Linux):

```sh
curl -fsSL https://raw.githubusercontent.com/abhishekbabu/croft/main/install.sh | sh
```

With Go:

```sh
go install github.com/abhishekbabu/croft/cmd/croft@latest
```

Or download a binary from the [releases page](https://github.com/abhishekbabu/croft/releases),
or build from source:

```sh
git clone https://github.com/abhishekbabu/croft
cd croft
go build ./cmd/croft
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

## Documentation

Full documentation is in [`docs/`](docs/README.md):

- [Getting started](docs/getting-started.md) — install and first worktree.
- [Command reference](docs/commands.md) — every command and flag.
- [Configuration](docs/configuration.md) — `croft.toml` and the machine config.
- [Providers](docs/providers.md) — the five swappable backends.
- [Agents](docs/agents.md) — agent runners and multi-agent coordination.
- [Migrating from an existing setup](docs/cutover.md) — incremental adoption.

See also [`examples/croft.toml`](examples/croft.toml) for a fully-wired config.

## Contributing

Issues and pull requests are welcome. Run `go build ./...` and `go test ./...`
before opening a PR; CI runs the same checks plus `golangci-lint`.

## License

[Apache-2.0](LICENSE)
