# Configuration

croft has two configuration layers and a runtime state directory.

## `croft.toml` — project config

Lives at the repository root, committed, shared by the whole team. Generate one
with `croft init`.

```toml
[project]
name = "myapp"                     # required

[worktree]
root        = "../worktrees"       # where checkouts land, relative to repo root
naming      = "myapp.{slug}"       # directory naming pattern; must contain {slug}
dev_command = "just dev"           # dev server command; {port} is substituted
copy_files  = [".env.local"]       # untracked files seeded into each worktree

[ports]
range    = "3000-3999"             # inclusive range each worktree draws from
services = ["api", "postgres"]     # each service gets a unique port per worktree

[providers]
multiplexer  = "tmux"              # none | tmux
infra        = "docker-compose"    # none | docker-compose
router       = "none"              # none | portless
stacker      = "none"              # none | graphite
coordination = "basic"             # basic | claude-agent-teams

[[agents]]
name   = "claude"
runner = "claude"                  # claude | codex | exec

[[agents]]
name    = "gemini"
runner  = "exec"
command = ["gemini", "--cwd", "{dir}"]   # required for the exec runner

[hooks]
post_create = ["docker compose up -d --wait"]
pre_remove  = ["docker compose down -v"]
```

Every `providers` slot has a `none` (or `basic`) default, so a repo with just
git and Docker works with zero external tools. The `{slug}` placeholder is the
worktree's stable, path-derived identity — it never changes when git HEAD moves.

## `~/.config/croft/config.toml` — machine config

Per-machine, gitignored. Everything that would otherwise be hardcoded to one
developer's setup. The file is optional.

```toml
[bins]
claude = "/opt/homebrew/bin/claude"   # tool -> absolute path; unset = resolve from PATH
codex  = "/opt/homebrew/bin/codex"

[defaults]
agent  = "claude"
effort = "xhigh"

[aws]
sso_profile = "myapp-dev"
```

`XDG_CONFIG_HOME` is honored; the default is `~/.config`.

## Runtime state

croft stores its registry under the XDG data directory —
`~/.local/share/croft/<project>/` (or `$XDG_DATA_HOME/croft/<project>/`):

- `registry.json` — the worktree registry (slug, branch, path, ports, URL).
- `peers/` — coordinated peer records and mailboxes.

State is plain JSON, one concern per file: no daemon, no database. Safe to
inspect, and `croft doctor` reconciles it against reality.

## See also

- [Providers](providers.md) — what each `[providers]` backend does.
- [Agents](agents.md) — the `[[agents]]` block and the runners behind it.
- [Command reference](commands.md) — every command and flag.
