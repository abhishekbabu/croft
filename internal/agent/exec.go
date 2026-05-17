package agent

import (
	"fmt"
	"strings"
)

// ExecRunner is the generic runner for any CLI agent (Gemini, Aider, …). Its
// argv template may contain {dir} and {prompt} placeholders, substituted at
// launch time.
type ExecRunner struct {
	name string
	argv []string
}

// Name returns the configured agent name.
func (r *ExecRunner) Name() string { return r.name }

// Launch builds the invocation by expanding placeholders in the argv template.
func (r *ExecRunner) Launch(spec Spec) (Invocation, error) {
	if len(r.argv) == 0 {
		return Invocation{}, fmt.Errorf("exec runner %q: empty command", r.name)
	}
	expanded := make([]string, len(r.argv))
	for i, a := range r.argv {
		a = strings.ReplaceAll(a, "{dir}", spec.Dir)
		a = strings.ReplaceAll(a, "{prompt}", spec.Prompt)
		expanded[i] = a
	}
	return Invocation{Path: expanded[0], Args: expanded[1:], Env: spec.Env, Dir: spec.Dir}, nil
}

// Resume re-launches the agent: a generic CLI agent has no standard resume
// protocol, so croft cannot do better than a fresh launch.
func (r *ExecRunner) Resume(_ string, spec Spec) (Invocation, error) {
	return r.Launch(spec)
}
