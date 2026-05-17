package agent

import "fmt"

// CodexRunner builds invocations for the OpenAI Codex CLI.
type CodexRunner struct {
	bin string
}

// Name returns "codex".
func (r *CodexRunner) Name() string { return "codex" }

// Launch builds the invocation that starts a Codex session. Headless launches
// use `codex exec --json`; the session id then arrives in the thread.started
// event — croft never scrapes rollout files (PLAN.md §2.5).
func (r *CodexRunner) Launch(spec Spec) (Invocation, error) {
	args := codexGlobalFlags(spec)
	if spec.Headless {
		args = append(args, "exec", "--json")
	}
	if spec.Prompt != "" {
		args = append(args, spec.Prompt)
	}
	return Invocation{Path: r.bin, Args: args, Env: spec.Env, Dir: spec.Dir}, nil
}

// Resume builds the invocation that resumes a Codex session by id.
func (r *CodexRunner) Resume(sessionID string, spec Spec) (Invocation, error) {
	if sessionID == "" {
		return Invocation{}, fmt.Errorf("codex resume: empty session id")
	}
	args := codexGlobalFlags(spec)
	if spec.Headless {
		args = append(args, "exec", "resume", sessionID, "--json")
	} else {
		args = append(args, "resume", sessionID)
	}
	if spec.Prompt != "" {
		args = append(args, spec.Prompt)
	}
	return Invocation{Path: r.bin, Args: args, Env: spec.Env, Dir: spec.Dir}, nil
}

// codexGlobalFlags renders the global flags (profile, model) that precede a
// Codex subcommand.
func codexGlobalFlags(spec Spec) []string {
	var args []string
	if spec.Profile != "" {
		args = append(args, "--profile", spec.Profile)
	}
	if spec.Model != "" {
		args = append(args, "-m", spec.Model)
	}
	return args
}
