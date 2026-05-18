package provider

// Infra manages a worktree's container stack.
type Infra interface {
	// Up brings the worktree's container stack online.
	Up(wt Worktree) error
	// Down tears the worktree's container stack down.
	Down(wt Worktree) error
	// Status reports whether the stack is running.
	Status(wt Worktree) (InfraState, error)
}

// NoneInfra is the no-op infra provider: worktrees have no container stack.
type NoneInfra struct{}

// Up does nothing.
func (NoneInfra) Up(Worktree) error { return nil }

// Down does nothing.
func (NoneInfra) Down(Worktree) error { return nil }

// Status always reports the stack as not running.
func (NoneInfra) Status(Worktree) (InfraState, error) {
	return InfraState{Up: false, Detail: "no infra provider"}, nil
}
