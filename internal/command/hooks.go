package command

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// runHooks executes a list of shell hook commands in dir with env exported.
// Each command runs via `sh -c`; the first failure stops the sequence.
func runHooks(label string, cmds []string, dir string, env map[string]string, out io.Writer) error {
	if len(cmds) == 0 {
		return nil
	}
	environ := os.Environ()
	for k, v := range env {
		environ = append(environ, k+"="+v)
	}
	for _, c := range cmds {
		fmt.Fprintf(out, "  hook (%s): %s\n", label, c)
		cmd := exec.Command("sh", "-c", c)
		cmd.Dir = dir
		cmd.Env = environ
		cmd.Stdout = out
		cmd.Stderr = out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s hook failed (%s): %w", label, c, err)
		}
	}
	return nil
}
