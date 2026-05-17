package provider

// Router gives a worktree a stable URL / route.
type Router interface {
	// Register routes traffic to the worktree and returns its URL.
	Register(wt Worktree) (url string, err error)
	// Release removes the worktree's route.
	Release(wt Worktree) error
}

// NoneRouter is the no-op router: worktrees get no managed route.
type NoneRouter struct{}

// Register returns an empty URL.
func (NoneRouter) Register(Worktree) (string, error) { return "", nil }

// Release does nothing.
func (NoneRouter) Release(Worktree) error { return nil }
