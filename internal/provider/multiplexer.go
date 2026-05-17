package provider

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// Multiplexer manages a terminal session per worktree.
type Multiplexer interface {
	// CreateSession starts a detached session named name, rooted at dir, with
	// env exported. Creating an existing session is a no-op.
	CreateSession(name, dir string, env map[string]string) error
	// RunWindow runs argv in a window of the session — the entry point for
	// launching an agent into a worktree.
	RunWindow(name, window, dir string, env map[string]string, argv []string) error
	// Attach connects the current terminal to the session.
	Attach(name string) error
	// Kill terminates the session. Killing an absent session is a no-op.
	Kill(name string) error
	// CapturePane returns the last n lines of a session window.
	CapturePane(name, window string, lines int) (string, error)
}

// NoneMultiplexer is the no-op multiplexer: worktrees have no managed session.
type NoneMultiplexer struct{}

// CreateSession does nothing.
func (NoneMultiplexer) CreateSession(string, string, map[string]string) error { return nil }

// RunWindow runs argv in the foreground, attached to the current terminal,
// blocking until it exits — with no multiplexer there is nowhere else to put
// an agent.
func (NoneMultiplexer) RunWindow(_, _, dir string, env map[string]string, argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("RunWindow: empty argv")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), envSlice(env)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// Attach does nothing.
func (NoneMultiplexer) Attach(string) error { return nil }

// Kill does nothing.
func (NoneMultiplexer) Kill(string) error { return nil }

// CapturePane returns an empty string.
func (NoneMultiplexer) CapturePane(string, string, int) (string, error) { return "", nil }

// TmuxMultiplexer drives sessions with tmux.
type TmuxMultiplexer struct {
	bin string
}

// NewTmuxMultiplexer returns a tmux-backed multiplexer. An empty bin resolves
// tmux from PATH.
func NewTmuxMultiplexer(bin string) *TmuxMultiplexer {
	if bin == "" {
		bin = "tmux"
	}
	return &TmuxMultiplexer{bin: bin}
}

// hasSession reports whether a tmux session with the given name exists.
func (t *TmuxMultiplexer) hasSession(name string) bool {
	return exec.Command(t.bin, "has-session", "-t", name).Run() == nil
}

// CreateSession starts a detached tmux session, idempotently.
func (t *TmuxMultiplexer) CreateSession(name, dir string, env map[string]string) error {
	if t.hasSession(name) {
		return nil
	}
	args := []string{"new-session", "-d", "-s", name, "-c", dir}
	for k, v := range env {
		args = append(args, "-e", k+"="+v)
	}
	_, err := run(t.bin, "", nil, args...)
	return err
}

// RunWindow opens a new tmux window in the session and runs argv there.
func (t *TmuxMultiplexer) RunWindow(name, window, dir string, env map[string]string, argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("RunWindow: empty argv")
	}
	args := []string{"new-window", "-t", name, "-c", dir}
	if window != "" {
		args = append(args, "-n", window)
	}
	for k, v := range env {
		args = append(args, "-e", k+"="+v)
	}
	args = append(args, argv...)
	_, err := run(t.bin, "", nil, args...)
	return err
}

// Attach connects the current terminal to the tmux session.
func (t *TmuxMultiplexer) Attach(name string) error {
	cmd := exec.Command(t.bin, "attach-session", "-t", name)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// Kill terminates the tmux session, idempotently.
func (t *TmuxMultiplexer) Kill(name string) error {
	if !t.hasSession(name) {
		return nil
	}
	_, err := run(t.bin, "", nil, "kill-session", "-t", name)
	return err
}

// CapturePane returns the last n lines of the named window.
func (t *TmuxMultiplexer) CapturePane(name, window string, lines int) (string, error) {
	target := name
	if window != "" {
		target = name + ":" + window
	}
	res, err := run(t.bin, "", nil,
		"capture-pane", "-p", "-t", target, "-S", "-"+strconv.Itoa(lines))
	return res.stdout, err
}
