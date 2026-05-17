package provider

import (
	"fmt"
	"os/exec"
	"strings"
)

// runResult is the captured outcome of an external command.
type runResult struct {
	stdout string
	stderr string
}

// run executes a command and returns its output. A non-zero exit becomes an
// error annotated with stderr, so provider failures are legible.
func run(name string, dir string, env []string, args ...string) (runResult, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = env
	}
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	res := runResult{stdout: out.String(), stderr: errb.String()}
	if err != nil {
		msg := strings.TrimSpace(res.stderr)
		if msg == "" {
			msg = err.Error()
		}
		return res, fmt.Errorf("%s %s: %s", name, strings.Join(args, " "), msg)
	}
	return res, nil
}

// available reports whether a binary is resolvable on PATH (or as an absolute
// path).
func available(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}
