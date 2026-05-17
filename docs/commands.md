# Command reference

Every croft command. Run `croft <command> --help` for the same flag summary on
the command line.

Global flags:

- `--help`, `-h` — help for croft or any subcommand.
- `--version`, `-v` — print the croft version.

All commands except `init` require a `croft.toml` at the repository root and
are run from anywhere inside the repository.

---

## `croft init`

```
croft init [--force] [--yes]
```

Scaffold a `croft.toml` at the repository root.

- By default `init` prompts interactively for the project name, worktree root,
  dev command, and providers.
- `--yes`, `-y` — accept defaults without prompting.
- `--force` — overwrite an existing `croft.toml`.

Must be run inside a git repository. The generated file is commented and
team-shareable. See [Configuration](configuration.md).

---

## `croft new`

```
croft new <branch> [--from <start-point>] [--agent <name>]
```

Create a fully isolated environment for `<branch>`. In order, `new`:

1. creates the git worktree (a new branch, or checks out an existing one),
2. allocates a unique port per declared service,
3. records the worktree in the registry,
4. creates the terminal session,
5. brings the container stack up,
6. registers the route,
7. runs `post_create` hooks,
8. starts the dev server, and
9. launches the agent, if `--agent` is given.

Flags:

- `--from <start-point>` — start point for a newly created branch (default:
  current HEAD). Ignored when the branch already exists.
- `--agent <name>` — launch a configured agent into the worktree's session.

`new` is **idempotent**: re-running it on an existing worktree re-converges the
environment (session, stack, route, dev server) without changing the port
allocation. It is **transactional**: if recording the worktree fails, the git
worktree is rolled back.

Steps 4–9 depend on the configured providers; with the `none` defaults `new`
creates only the git worktree.

---

## `croft ls`

```
croft ls
```

List every croft-managed worktree as a table: `SLUG`, `BRANCH`, `STATUS`,
`PORTS`, `PATH`.

`STATUS` is derived live:

| Status | Meaning |
|--------|---------|
| `-` | No agent launched. |
| `working` | An agent is running in the worktree's session. |
| `done` | An agent was launched but its window has exited. |
| `rebase` | The worktree has a rebase in progress. |
| `missing` | The registry entry's directory no longer exists. |

---

## `croft status`

```
croft status <branch>
```

Show detail for one worktree: slug, branch, path, URL (if routed), whether the
directory exists, derived status, ports, and creation time.

---

## `croft sync`

```
croft sync [branch] [--prune]
```

Rebase branch stacks against the trunk via the stacker provider. With no
argument it syncs every registered worktree; with `[branch]` it syncs one.

Per worktree, `sync`:

1. refuses if the worktree is mid-rebase (finish with `gt continue` or
   `git rebase --abort` first),
2. auto-stashes any uncommitted work,
3. rebases the stack,
4. pops the stash — a failed pop is reported loudly, since silent stash loss is
   the worst outcome,
5. applies the **stack-resolved teardown gate**: if every branch in the stack
   is merged or closed, the worktree is reported as fully resolved.

- `--prune` — tear down worktrees whose stack is fully resolved, instead of
  just reporting them.

With the `none` stacker, `sync` is a graceful no-op per worktree.

---

## `croft rm`

```
croft rm <branch> [--force]
```

Tear a worktree and its environment down, in reverse of `new`: `pre_remove`
hooks, container stack down, route released, session killed, git worktree
removed, stale git entries pruned, registry record cleared.

- `--force`, `-f` — remove even when the worktree has uncommitted changes.

`rm` is **idempotent**: a worktree that is already partly gone is cleaned up
the rest of the way without error.

---

## `croft doctor`

```
croft doctor [--fix]
```

Reconcile croft state against reality. It runs four checks:

1. **Stale registry entries** — a registry record whose directory is gone.
2. **Stale git worktree entries** — `git worktree` admin entries with no
   directory.
3. **Orphan directories** — directories under the worktree root that match the
   naming pattern but have no registry entry.
4. **Leaked container stacks** — `docker compose` projects named for this
   project with no matching worktree.

By default `doctor` reports findings. With `--fix` it applies an idempotent
repair for each. Checks degrade gracefully when a tool (git, docker) is absent.

---

## `croft spawn`

```
croft spawn <name> --agent <agent> [--role <role>] [--dir <dir>]
```

Spawn a named coordinated agent (a "peer") via the coordination provider.

- `--agent <agent>` — the configured agent to run as this peer (required).
- `--role <role>` — role text appended to the agent's system prompt.
- `--dir <dir>` — working directory for the peer (default: current directory).

See [Agents](agents.md) for how peers and coordination work.

---

## `croft fleet`

```
croft fleet status
croft fleet msg <peer> <message>
```

Cross-worktree coordination. Peers share one per-project state directory, so
the fleet spans every worktree of the project.

- `fleet status` — list every coordinated peer (`PEER`, `AGENT`, `STATUS`).
- `fleet msg <peer> <message>` — append a message to a peer's mailbox.
