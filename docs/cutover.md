# Migrating to croft from an existing worktree setup

croft was extracted from a personal shell-based worktree harness. If you have
your own scripts managing worktrees, containers, and ports, you can move to
croft incrementally — there is no flag day.

## Parallel-run model

croft drives the *same* worktrees your existing tooling already manages. As
long as `croft.toml` points its providers at the tools you already use
(`tmux`, `docker-compose`, `portless`, `graphite`), both can operate against
one repository during the transition.

1. **Write `croft.toml`.** Run `croft init` at the repo root and set the
   `providers` block to match your current stack. See `examples/croft.toml`
   for a fully-wired example (tmux + portless + graphite).

2. **Adopt verb by verb.** croft's commands map onto the usual worktree
   operations:

   | croft         | typical harness equivalent              |
   |---------------|-----------------------------------------|
   | `croft new`   | create worktree + infra + ports + route |
   | `croft ls`    | worktree status dashboard               |
   | `croft status`| per-worktree detail                     |
   | `croft sync`  | stack rebase + merged-PR teardown gate   |
   | `croft rm`    | worktree teardown                        |
   | `croft doctor`| orphan / drift cleanup                   |
   | `croft spawn` / `croft fleet` | multi-agent spawn + dispatch |

   Prove one verb against a real worktree, then retire the equivalent script.

3. **Trust `sync` and `doctor` last.** They have the highest blast radius —
   `sync` rebases stacks, `doctor` deletes things. Keep your existing versions
   until croft's have run clean for a while.

At every step either tool can manage the worktrees, because croft stores its
state separately (under the XDG data dir) and wraps — never reimplements — the
underlying tools (`git worktree`, `tmux`, `docker compose`, `gt`, `gh`).

## Verifying

After pointing `croft.toml` at your stack:

- `croft ls` should list the worktrees you already have (run `croft doctor` to
  adopt any it does not yet know about).
- `croft new <branch>` should produce a worktree indistinguishable from one
  your old tooling makes — same directory, same container project namespace,
  same port scheme.
- `croft rm <branch>` should tear it down completely; `croft doctor` should
  then report "all clear".
