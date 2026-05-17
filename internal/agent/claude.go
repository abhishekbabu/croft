package agent

import "fmt"

// ClaudeRunner builds invocations for the Claude Code CLI.
type ClaudeRunner struct {
	bin string
}

// Name returns "claude".
func (r *ClaudeRunner) Name() string { return "claude" }

// Launch builds the invocation that starts a Claude session. Interactive
// launches pin the session with --session-id; headless launches add
// `-p --output-format json --bare` for a reproducible, parseable run.
func (r *ClaudeRunner) Launch(spec Spec) (Invocation, error) {
	var args []string
	if spec.Headless {
		args = append(args, "-p", "--output-format", "json", "--bare")
	}
	if spec.SessionID != "" {
		args = append(args, "--session-id", spec.SessionID)
	}
	if spec.Model != "" {
		args = append(args, "--model", spec.Model)
	}
	if spec.AppendSystemPrompt != "" {
		args = append(args, "--append-system-prompt", spec.AppendSystemPrompt)
	}
	if spec.Prompt != "" {
		args = append(args, spec.Prompt)
	}
	return Invocation{Path: r.bin, Args: args, Env: spec.Env, Dir: spec.Dir}, nil
}

// Resume builds the invocation that resumes a Claude session by id.
func (r *ClaudeRunner) Resume(sessionID string, spec Spec) (Invocation, error) {
	if sessionID == "" {
		return Invocation{}, fmt.Errorf("claude resume: empty session id")
	}
	var args []string
	if spec.Headless {
		args = append(args, "-p", "--output-format", "json")
	}
	args = append(args, "--resume", sessionID)
	if spec.Prompt != "" {
		args = append(args, spec.Prompt)
	}
	return Invocation{Path: r.bin, Args: args, Env: spec.Env, Dir: spec.Dir}, nil
}
