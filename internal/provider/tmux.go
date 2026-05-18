package provider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/abhishekbabu/croft/internal/sh"
)

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

// Managed reports true: tmux hosts real sessions.
func (t *TmuxMultiplexer) Managed() bool { return true }

// HasWindow reports whether a window of the given name exists in the session.
func (t *TmuxMultiplexer) HasWindow(name, window string) bool {
	res, err := sh.Capture(t.bin, "", nil, "list-windows", "-t", name, "-F", "#{window_name}")
	if err != nil {
		return false
	}
	for _, w := range strings.Split(res, "\n") {
		if strings.TrimSpace(w) == window {
			return true
		}
	}
	return false
}

// hasSession reports whether a tmux session with the given name exists.
func (t *TmuxMultiplexer) hasSession(name string) bool {
	_, err := sh.Capture(t.bin, "", nil, "has-session", "-t", name)
	return err == nil
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
	_, err := sh.Capture(t.bin, "", nil, args...)
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
	_, err := sh.Capture(t.bin, "", nil, args...)
	return err
}

// Attach connects the current terminal to the tmux session.
func (t *TmuxMultiplexer) Attach(name string) error {
	return sh.Attach(t.bin, "", nil, "attach-session", "-t", name)
}

// Kill terminates the tmux session, idempotently.
func (t *TmuxMultiplexer) Kill(name string) error {
	if !t.hasSession(name) {
		return nil
	}
	_, err := sh.Capture(t.bin, "", nil, "kill-session", "-t", name)
	return err
}

// CapturePane returns the last n lines of the named window.
func (t *TmuxMultiplexer) CapturePane(name, window string, lines int) (string, error) {
	target := name
	if window != "" {
		target = name + ":" + window
	}
	res, err := sh.Capture(t.bin, "", nil,
		"capture-pane", "-p", "-t", target, "-S", "-"+strconv.Itoa(lines))
	return res, err
}
