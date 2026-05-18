package agent

import (
	"regexp"
	"strings"
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
)

var uuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestDeterministicSessionID(t *testing.T) {
	a := DeterministicSessionID("demo:feat:claude")
	b := DeterministicSessionID("demo:feat:claude")
	c := DeterministicSessionID("demo:other:claude")
	if a != b {
		t.Errorf("session id not stable: %q vs %q", a, b)
	}
	if a == c {
		t.Error("different keys produced the same session id")
	}
	if !uuidV4.MatchString(a) {
		t.Errorf("session id %q is not a v4 UUID", a)
	}
}

func TestClaudeLaunch(t *testing.T) {
	r := &ClaudeRunner{bin: "claude"}

	inv, err := r.Launch(Spec{Dir: "/wt", SessionID: "sid-1"})
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if inv.Path != "claude" || inv.Dir != "/wt" {
		t.Errorf("invocation = %+v", inv)
	}
	if !contains(inv.Args, "--session-id") || !contains(inv.Args, "sid-1") {
		t.Errorf("interactive launch missing session id: %v", inv.Args)
	}
	if contains(inv.Args, "-p") {
		t.Errorf("interactive launch should not be headless: %v", inv.Args)
	}

	headless, _ := r.Launch(Spec{Headless: true})
	if !contains(headless.Args, "-p") || !contains(headless.Args, "--bare") {
		t.Errorf("headless launch missing -p/--bare: %v", headless.Args)
	}
}

func TestClaudeResume(t *testing.T) {
	r := &ClaudeRunner{bin: "claude"}
	inv, err := r.Resume("sid-9", Spec{})
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if !contains(inv.Args, "--resume") || !contains(inv.Args, "sid-9") {
		t.Errorf("resume args = %v", inv.Args)
	}
	if _, err := r.Resume("", Spec{}); err == nil {
		t.Error("resume with empty id should error")
	}
}

func TestCodexLaunch(t *testing.T) {
	r := &CodexRunner{bin: "codex"}

	inv, _ := r.Launch(Spec{Headless: true, Profile: "dev", Prompt: "do it"})
	if !contains(inv.Args, "exec") || !contains(inv.Args, "--json") {
		t.Errorf("headless codex launch missing exec --json: %v", inv.Args)
	}
	if !contains(inv.Args, "--profile") || !contains(inv.Args, "dev") {
		t.Errorf("codex launch missing profile: %v", inv.Args)
	}

	interactive, _ := r.Launch(Spec{})
	if contains(interactive.Args, "exec") {
		t.Errorf("interactive codex launch should not use exec: %v", interactive.Args)
	}

	resume, _ := r.Resume("th-1", Spec{Headless: true})
	if !contains(resume.Args, "resume") || !contains(resume.Args, "th-1") {
		t.Errorf("codex resume args = %v", resume.Args)
	}
}

func TestExecRunnerSubstitutes(t *testing.T) {
	r := &ExecRunner{name: "gemini", argv: []string{"gemini", "--cwd", "{dir}", "{prompt}"}}
	inv, err := r.Launch(Spec{Dir: "/wt/demo", Prompt: "fix bug"})
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if inv.Path != "gemini" {
		t.Errorf("path = %q", inv.Path)
	}
	joined := strings.Join(inv.Args, " ")
	if !strings.Contains(joined, "/wt/demo") || !strings.Contains(joined, "fix bug") {
		t.Errorf("placeholders not substituted: %v", inv.Args)
	}
}

func TestNewFactory(t *testing.T) {
	m := config.MachineConfig{}
	for _, runner := range []config.AgentRunner{config.RunnerClaude, config.RunnerCodex} {
		got, err := New(config.AgentConfig{Name: string(runner), Runner: runner}, m)
		if err != nil {
			t.Fatalf("New(%s): %v", runner, err)
		}
		if got.Name() != string(runner) {
			t.Errorf("New(%s).Name() = %q", runner, got.Name())
		}
	}
	if _, err := New(config.AgentConfig{Name: "x", Runner: "exec"}, m); err == nil {
		t.Error("exec runner without a command should error")
	}
	if _, err := New(config.AgentConfig{Name: "x", Runner: "bogus"}, m); err == nil {
		t.Error("unknown runner should error")
	}
}

// contains reports whether set includes v.
func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}
