package provider

import (
	"testing"

	"github.com/abhishekbabu/croft/internal/config"
)

func TestEnv(t *testing.T) {
	env := Env(Worktree{
		Slug:  "feat",
		Path:  "/wt/demo.feat",
		Ports: map[string]int{"api": 3000, "postgres": 3001},
	})
	if env["CROFT_SLUG"] != "feat" {
		t.Errorf("CROFT_SLUG = %q", env["CROFT_SLUG"])
	}
	if env["COMPOSE_PROJECT_NAME"] != "croft-feat" {
		t.Errorf("COMPOSE_PROJECT_NAME = %q", env["COMPOSE_PROJECT_NAME"])
	}
	if env["API_PORT"] != "3000" || env["POSTGRES_PORT"] != "3001" {
		t.Errorf("port env wrong: %v", env)
	}
}

func TestNewSelectsImplementations(t *testing.T) {
	set, err := New(config.ProvidersSection{
		Multiplexer:  "tmux",
		Infra:        "docker-compose",
		Router:       "none",
		Stacker:      "none",
		Coordination: "basic",
	}, config.MachineConfig{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := set.Multiplexer.(*TmuxMultiplexer); !ok {
		t.Errorf("multiplexer = %T, want *TmuxMultiplexer", set.Multiplexer)
	}
	if _, ok := set.Infra.(*ComposeInfra); !ok {
		t.Errorf("infra = %T, want *ComposeInfra", set.Infra)
	}
}

func TestNewDefaultsToNoOp(t *testing.T) {
	set, err := New(config.ProvidersSection{}, config.MachineConfig{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := set.Multiplexer.(NoneMultiplexer); !ok {
		t.Errorf("multiplexer = %T, want NoneMultiplexer", set.Multiplexer)
	}
	if _, ok := set.Infra.(NoneInfra); !ok {
		t.Errorf("infra = %T, want NoneInfra", set.Infra)
	}
}

func TestNewRejectsUnknownProvider(t *testing.T) {
	if _, err := New(config.ProvidersSection{Infra: "nomad"}, config.MachineConfig{}); err == nil {
		t.Error("expected error for unknown infra provider")
	}
}

func TestNoneProvidersAreInert(t *testing.T) {
	wt := Worktree{Slug: "x", Path: "/tmp/x"}
	if err := (NoneInfra{}).Up(wt); err != nil {
		t.Errorf("NoneInfra.Up: %v", err)
	}
	st, _ := (NoneInfra{}).Status(wt)
	if st.Up {
		t.Error("NoneInfra.Status should report down")
	}
	if err := (NoneMultiplexer{}).CreateSession("x", "/tmp", nil); err != nil {
		t.Errorf("NoneMultiplexer.CreateSession: %v", err)
	}
	if url, _ := (NoneRouter{}).Register(wt); url != "" {
		t.Errorf("NoneRouter.Register url = %q, want empty", url)
	}
	if ok, _ := (NoneStacker{}).AllResolved(wt); ok {
		t.Error("NoneStacker.AllResolved must be false")
	}
}

func TestTmuxLifecycle(t *testing.T) {
	if !available("tmux") {
		t.Skip("tmux not installed")
	}
	tm := NewTmuxMultiplexer("")
	const name = "croft-test-session"
	_ = tm.Kill(name) // clean any leftover
	t.Cleanup(func() { _ = tm.Kill(name) })

	if err := tm.CreateSession(name, t.TempDir(), map[string]string{"FOO": "bar"}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	// CreateSession is idempotent.
	if err := tm.CreateSession(name, t.TempDir(), nil); err != nil {
		t.Errorf("CreateSession (repeat): %v", err)
	}
	if !tm.hasSession(name) {
		t.Error("session should exist after CreateSession")
	}
	if err := tm.Kill(name); err != nil {
		t.Errorf("Kill: %v", err)
	}
	// Kill is idempotent.
	if err := tm.Kill(name); err != nil {
		t.Errorf("Kill (repeat): %v", err)
	}
}
