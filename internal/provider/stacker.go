package provider

// Stacker syncs a worktree's branch stack against the trunk and reports stack
// resolution (used by `croft sync` and the teardown gate).
type Stacker interface {
	// Sync rebases the worktree's stack onto the trunk.
	Sync(wt Worktree) (StackState, error)
	// StackBranches lists every branch in the worktree's stack.
	StackBranches(wt Worktree) ([]string, error)
	// AllResolved reports whether every branch in the stack is merged/closed.
	AllResolved(wt Worktree) (bool, error)
}

// NoneStacker is the no-op stacker: croft has no stack awareness.
type NoneStacker struct{}

// Sync does nothing.
func (NoneStacker) Sync(Worktree) (StackState, error) {
	return StackState{Detail: "no stacker provider"}, nil
}

// StackBranches returns no branches.
func (NoneStacker) StackBranches(Worktree) ([]string, error) { return nil, nil }

// AllResolved reports false: with no stacker, resolution is unknown, so the
// teardown gate must never fire automatically.
func (NoneStacker) AllResolved(Worktree) (bool, error) { return false, nil }
