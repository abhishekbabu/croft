# Getting started

## Install

Install script (Linux and macOS):

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

Verify:

```sh
croft --version
```

## Initialize a repository

From inside a git repository, scaffold a `croft.toml`:

```sh
croft init
```

`init` prompts for the project name, where worktrees should land, the dev
server command, and which providers to use. Pass `--yes` to accept defaults
without prompting, or `--force` to overwrite an existing `croft.toml`.

The result is a committed, team-shared `croft.toml` at the repository root. See
[Configuration](configuration.md) for every field.

## Create your first worktree

```sh
croft new my-feature
```

This creates a git worktree for the `my-feature` branch and, depending on the
configured providers:

- allocates a unique port for each declared service,
- brings up an isolated container stack,
- creates a terminal session,
- registers a route, and
- starts the dev server.

`croft new` is idempotent — running it again on an existing worktree
re-converges the environment instead of failing.

To launch an AI coding agent into the worktree at the same time:

```sh
croft new my-feature --agent claude
```

## Inspect and tear down

```sh
croft ls                  # all worktrees and their status
croft status my-feature   # detail for one worktree
croft rm my-feature       # tear the worktree and its environment down
```

## A minimal, zero-dependency setup

Every provider defaults to a no-op, so croft is useful with just git. A minimal
`croft.toml`:

```toml
[project]
name = "myapp"
```

`croft new` then creates only the git worktree — no containers, no session, no
route. Add providers as you need them; see [Providers](providers.md).

## Next steps

- [Command reference](commands.md) — every command in detail.
- [Configuration](configuration.md) — the full `croft.toml` schema.
- [Agents](agents.md) — launching and coordinating AI agents.
