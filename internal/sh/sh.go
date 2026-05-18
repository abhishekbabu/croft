// Package sh runs external commands with consistent error handling and
// bounded execution. croft shells out heavily (git, tmux, cmux, docker,
// gt, gh, agents); routing every call through here gives uniform error
// messages and a single place that enforces timeouts and signal handling.
package sh

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DefaultTimeout bounds how long a single captured or streamed command may
// run before it is killed. It is a var so tests and callers can tune it.
var DefaultTimeout = 10 * time.Minute

// baseCtx is the parent of every command's context. SetBaseContext installs
// one cancelled on SIGINT so Ctrl-C kills in-flight child processes.
var baseCtx = context.Background()

// SetBaseContext installs the base context every command derives from. Call
// once at startup, before any command runs.
func SetBaseContext(ctx context.Context) { baseCtx = ctx }

// Capture runs name+args in dir (empty = current directory) with env (nil =
// inherit the parent environment) and returns stdout. A non-zero exit becomes
// an error annotated with stderr; exceeding DefaultTimeout becomes a timeout
// error.
func Capture(name, dir string, env []string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(baseCtx, DefaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	var out, errb strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errb

	if err := cmd.Run(); err != nil {
		return out.String(), annotate(ctx, name, args, errb.String(), err)
	}
	return out.String(), nil
}

// StreamTo runs name+args with stdout and stderr both written to w. Used for
// hooks, where the command's output belongs in croft's own output.
func StreamTo(w io.Writer, name, dir string, env []string, args ...string) error {
	ctx, cancel := context.WithTimeout(baseCtx, DefaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout, cmd.Stderr = w, w
	if err := cmd.Run(); err != nil {
		return annotate(ctx, name, args, "", err)
	}
	return nil
}

// Attach runs name+args wired to the current terminal. It is not bounded by a
// timeout — it is for interactive processes (an attached session, a foreground
// agent) — but a cancelled base context still terminates it.
func Attach(name, dir string, env []string, args ...string) error {
	cmd := exec.CommandContext(baseCtx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// Look reports whether bin resolves on PATH (or as an absolute path).
func Look(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

// annotate turns a failed command into a legible error: a timeout when the
// context deadline was hit, otherwise the trimmed stderr.
func annotate(ctx context.Context, name string, args []string, stderr string, runErr error) error {
	cmdline := strings.TrimSpace(name + " " + strings.Join(args, " "))
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("%s: timed out after %s", cmdline, DefaultTimeout)
	}
	if msg := strings.TrimSpace(stderr); msg != "" {
		return fmt.Errorf("%s: %s", cmdline, msg)
	}
	return fmt.Errorf("%s: %w", cmdline, runErr)
}
