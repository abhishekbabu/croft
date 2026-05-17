package state

import (
	"testing"
	"time"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := OpenAt(t.TempDir())
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	return s
}

func TestLoadEmptyRegistry(t *testing.T) {
	r, err := newStore(t).Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if r.Worktrees == nil || len(r.Worktrees) != 0 {
		t.Errorf("fresh registry should be empty and non-nil, got %+v", r)
	}
}

func TestPutGetDelete(t *testing.T) {
	s := newStore(t)
	wt := Worktree{
		Slug:    "feat",
		Branch:  "my-feature",
		Path:    "/wt/demo.feat",
		Ports:   map[string]int{"api": 3000},
		Status:  StatusWorking,
		Created: time.Now().UTC().Truncate(time.Second),
	}
	if err := s.Put(wt); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, found, err := s.Get("feat")
	if err != nil || !found {
		t.Fatalf("Get: found=%v err=%v", found, err)
	}
	if got.Branch != "my-feature" || got.Ports["api"] != 3000 || got.Status != StatusWorking {
		t.Errorf("Get returned %+v", got)
	}

	if err := s.Delete("feat"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, found, _ := s.Get("feat"); found {
		t.Error("record should be gone after Delete")
	}
	// Deleting an absent slug must be a no-op, not an error.
	if err := s.Delete("feat"); err != nil {
		t.Errorf("Delete of absent slug should be idempotent: %v", err)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	s := newStore(t)
	in := Registry{Worktrees: map[string]Worktree{
		"a": {Slug: "a", Branch: "ba"},
		"b": {Slug: "b", Branch: "bb"},
	}}
	if err := s.Save(in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	out, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(out.Worktrees) != 2 || out.Worktrees["b"].Branch != "bb" {
		t.Errorf("round trip mismatch: %+v", out)
	}

	// Reopening the same directory must see persisted state.
	s2, err := OpenAt(s.Dir())
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if r, _ := s2.Load(); len(r.Worktrees) != 2 {
		t.Errorf("state did not persist across Store instances")
	}
}
