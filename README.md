# croft

> Worktrees that actually work.

`croft` gives every git branch its own fully isolated, working development
environment — its own checkout, container stack, ports, route, dev server, and
(optionally) an AI coding agent — created and torn down in seconds.

## Why

Lots of tools spawn AI agents into git worktrees. Nearly all of them punt on
**runtime isolation**: worktrees collide on ports, share one `docker compose`
namespace, and write to the same dev database. `croft` is the isolation
substrate underneath — it makes worktrees *actually work* as the runtime each
branch (and each agent) lives in.

## Status

Early development. Building Phase 1 milestone by milestone — see `PLAN.md` for
the full specification and roadmap.

- [x] **M0** — Scaffold (module, CLI skeleton, CI)
- [ ] **M1** — Config layer + `croft init`
- [ ] **M2** — Worktree core (`new`, `ls`, `status`, `rm`)
- [ ] **M3** — Provider interfaces + isolation
- [ ] **M4** — Agent runners (Claude, Codex, generic)
- [ ] **M5** — Stacker + router (`sync`)
- [ ] **M6** — `doctor`
- [ ] **M7** — Coordination (`spawn`, `fleet`)
- [ ] **M8** — Cutover support
- [ ] **M9** — Release v0.1

## Build

```sh
go build ./...
go test ./...
```

## License

[Apache-2.0](LICENSE)
