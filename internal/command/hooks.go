package command

import (
	"fmt"
	"io"
	"os"

	"github.com/abhishekbabu/croft/internal/sh"
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
		if err := sh.StreamTo(out, "sh", dir, environ, "-c", c); err != nil {
			return fmt.Errorf("%s hook failed (%s): %w", label, c, err)
		}
	}
	return nil
}
