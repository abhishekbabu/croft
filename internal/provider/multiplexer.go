package provider

import (
	"fmt"
	"os"

	"github.com/abhishekbabu/croft/internal/sh"
)

// Multiplexer manages a terminal session per worktree.
type Multiplexer interface {
	// Managed reports whether the multiplexer can host long-running background
	// processes in a real session. The none multiplexer cannot.
	Managed() bool
	// CreateSession starts a detached session named name, rooted at dir, with
	// env exported. Creating an existing session is a no-op.
	CreateSession(name, dir string, env map[string]string) error
	// RunWindow runs argv in a window of the session — the entry point for
	// launching an agent or dev server into a worktree.
	RunWindow(name, window, dir string, env map[string]string, argv []string) error
	// HasWindow reports whether a window of the given name exists in the
	// session — used to keep RunWindow callers idempotent.
	HasWindow(name, window string) bool
	// Attach connects the current terminal to the session.
	Attach(name string) error
	// Kill terminates the session. Killing an absent session is a no-op.
	Kill(name string) error
	// CapturePane returns the last n lines of a session window.
	CapturePane(name, window string, lines int) (string, error)
}

// NoneMultiplexer is the no-op multiplexer: worktrees have no managed session.
type NoneMultiplexer struct{}

// Managed reports false: the none multiplexer hosts nothing.
func (NoneMultiplexer) Managed() bool { return false }

// CreateSession does nothing.
func (NoneMultiplexer) CreateSession(string, string, map[string]string) error { return nil }

// HasWindow always reports false.
func (NoneMultiplexer) HasWindow(string, string) bool { return false }

// RunWindow runs argv in the foreground, attached to the current terminal,
// blocking until it exits — with no multiplexer there is nowhere else to put
// an agent.
func (NoneMultiplexer) RunWindow(_, _, dir string, env map[string]string, argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("RunWindow: empty argv")
	}
	return sh.Attach(argv[0], dir, append(os.Environ(), envSlice(env)...), argv[1:]...)
}

// Attach does nothing.
func (NoneMultiplexer) Attach(string) error { return nil }

// Kill does nothing.
func (NoneMultiplexer) Kill(string) error { return nil }

// CapturePane returns an empty string.
func (NoneMultiplexer) CapturePane(string, string, int) (string, error) { return "", nil }
