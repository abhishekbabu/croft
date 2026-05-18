# Providers

A provider is a swappable backend for one concern. croft has five provider
slots; each is set in the `[providers]` block of `croft.toml`:

```toml
[providers]
multiplexer  = "tmux"
infra        = "docker-compose"
router       = "portless"
stacker      = "graphite"
coordination = "basic"
```

Every slot has a no-op (`none` / `basic`) default, so a repository with just
git — and Docker, if you want the infra slot — works with zero extra tools.
croft's core calls each provider through an interface and never branches on
which backend is active.

## Namespacing

Each worktree gets a stable namespace, `<project>-<slug>`, shared by its
container stack and its terminal session — so neither collides across
worktrees or across projects. croft also exports these environment variables
into the session and container stack:

| Variable | Value |
|----------|-------|
| `CROFT_SLUG` | the worktree's slug |
| `CROFT_WORKTREE` | the worktree's absolute path |
| `COMPOSE_PROJECT_NAME` | `<project>-<slug>` |
| `<SERVICE>_PORT` | the allocated port, one per declared service |

---

## Multiplexer

Manages a terminal session per worktree, and hosts the dev server and agents.

### `none` (default)

No managed session. The dev server command is printed for you to run yourself;
an agent launched with `--agent` runs in the foreground.

### `tmux`

A detached tmux session named `<project>-<slug>`. The dev server runs in a
`dev` window; an agent runs in an `agent` window. `croft new` is idempotent —
it will not open a second `dev` or `agent` window if one already exists.
Requires `tmux` on `PATH` (or a `[bins]` override).

### `cmux`

A cmux *workspace* is the session; a cmux *surface* is a window. cmux only
gives a surface a live terminal while it is rendered on screen, so croft must
be **run from inside a focused cmux terminal**. Each window is created by
focusing croft's own surface, splitting it (the split is live), running the
command there, then moving the live surface into the worktree's workspace.

If croft's surface cannot be focused — croft was not run from a cmux terminal —
the operation refuses with a clear message rather than creating a dead surface.
Requires `$CMUX_SURFACE_ID` (set automatically inside cmux terminals).

---

## Infra

Manages a worktree's container stack.

### `none` (default)

No container stack.

### `docker-compose`

Runs a per-worktree `docker compose` stack in its own project namespace
(`<project>-<slug>`), so stacks never collide. `croft new` runs
`docker compose up -d`; `croft rm` runs `docker compose down -v`. The
repository's compose file is read from the worktree checkout, and croft's
environment variables are exported so the compose file can reference
`${SERVICE_PORT}`. Requires `docker` on `PATH`.

---

## Router

Gives a worktree a stable URL.

### `none` (default)

No route.

### `portless`

Registers a static portless route per worktree service
(`portless alias <slug>-<service> <port>`) and records the worktree's URL.
`croft rm` releases the routes; release is best-effort, so a route that is
already gone never blocks teardown. Requires `portless` on `PATH`.

---

## Stacker

Syncs a worktree's branch stack against the trunk and reports stack resolution.

### `none` (default)

No stack awareness. `croft sync` is a graceful no-op, and the stack-resolved
teardown gate never fires.

### `graphite`

Drives branch stacks with the Graphite CLI (`gt`):

- **Sync** — `gt sync --no-interactive` rebases the stack. `--force` is
  deliberately omitted; deciding teardown is croft's job, not gt's.
- **Stack branches** — parsed from `gt log short -s`.
- **Resolution** — every stack branch is cross-checked against PR state with a
  single `gh pr list` for the whole repo (never N round-trips). A stack whose
  every branch is merged, closed, or has no PR is "fully resolved".

Requires `gt`; PR resolution additionally uses `gh` (degrading gracefully to
"not resolved" when `gh` is absent).

---

## Coordination

Spawns and coordinates named agents across worktrees. See [Agents](agents.md).

### `basic` (default)

File-based: a peer registry and mailbox under the project's shared state
directory, with agents hosted in a shared multiplexer session. Agent-agnostic —
works with any agent runner.

### `claude-agent-teams`

Currently uses the same file-based backend as `basic`. Native Claude Agent
Teams integration is a planned Phase 2 enhancement.
