# Agents

croft can launch AI coding agents into a worktree and coordinate several of
them. Agent *spawning* is a supporting feature — croft's focus is the isolation
layer underneath — but it is fully wired.

## Declaring agents

Agents are declared in `croft.toml` as repeated `[[agents]]` tables. Each has a
`name` (how you refer to it) and a `runner` (which CLI backs it):

```toml
[[agents]]
name   = "claude"
runner = "claude"

[[agents]]
name   = "codex"
runner = "codex"

[[agents]]
name    = "gemini"
runner  = "exec"
command = ["gemini", "--cwd", "{dir}"]
```

## Runners

A runner builds the process invocation for an agent CLI — croft never
reimplements an agent, it only shells out.

### `claude`

Invokes the Claude Code CLI. Interactive launches pin the session with
`--session-id` set to a deterministic UUID derived from the project, worktree,
and agent name — so re-launching the same agent reuses the same session.
Headless launches add `-p --output-format json --bare` for a reproducible,
parseable run.

### `codex`

Invokes the OpenAI Codex CLI. Headless launches use `codex exec --json`; the
session id arrives in the `thread.started` event (croft never scrapes rollout
files). Resume uses `codex exec resume <id>`.

### `exec`

A generic runner for any CLI agent (Gemini, Aider, …). The required `command`
array is the argv template; the placeholders `{dir}` (the worktree path) and
`{prompt}` are substituted at launch.

## Launching an agent into a worktree

```sh
croft new my-feature --agent claude
```

The agent runs in the worktree's session — an `agent` window with the `tmux`
multiplexer, or the foreground with the `none` multiplexer. Launching is
idempotent: with a managed multiplexer, croft will not start a second agent if
one is already running in the worktree.

The worktree's `ls`/`status` reflects the agent: `working` while its window is
alive, `done` once it exits.

## Coordination — spawning a fleet

Beyond a single per-worktree agent, croft can spawn named, coordinated agents
("peers") and dispatch messages between them. This is the `coordination`
provider; the default `basic` backend is file-based and agent-agnostic.

### Spawn a peer

```sh
croft spawn reviewer --agent claude --role "review the diff for security issues"
```

- `--agent` — the configured agent to run as this peer.
- `--role` — text appended to the agent's system prompt.
- `--dir` — the peer's working directory (default: current directory).

The peer's agent is launched into a shared `<project>-peers` session, and a
peer record is written to the project's state directory. croft exports
`CROFT_PEER` (and `CROFT_PEER_ROLE`, if set) into the peer's environment.

### Inspect and message the fleet

```sh
croft fleet status                 # list every peer: name, agent, status
croft fleet msg reviewer "ship it" # append a message to a peer's mailbox
```

Peers share one per-project state directory, so `croft fleet` spans every
worktree of the project — a peer spawned from one worktree is visible from
another.

### Backends

- **`basic`** — file-based peer registry and mailbox; works with any runner.
- **`claude-agent-teams`** — currently the same file-based backend; native
  Claude Agent Teams integration is a planned Phase 2 enhancement.
