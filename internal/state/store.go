// Package state persists croft's runtime state as plain JSON under an XDG data
// directory — no daemon, no database, one concern per file (PLAN.md §6.2).
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/abhishekbabu/croft/internal/env"
)

// Status is a worktree's agent-state marker (PLAN.md §2.3). The status
// dashboard renders these; the empty Status means "no agent".
type Status string

// Worktree agent-state markers. StatusWorking is the only value croft
// persists; StatusDone, StatusRebase, and StatusMissing are derived live by
// the command layer and never written to the registry.
const (
	StatusWorking Status = "working" // an agent is running in the worktree
	StatusDone    Status = "done"    // the agent's window has exited (derived)
	StatusRebase  Status = "rebase"  // a rebase is in progress (derived)
	StatusMissing Status = "missing" // the worktree directory is gone (derived)
)

// Worktree is one registry record describing a croft-managed checkout.
type Worktree struct {
	Slug    string         `json:"slug"`
	Branch  string         `json:"branch"`
	Path    string         `json:"path"`
	Ports   map[string]int `json:"ports,omitempty"`
	URL     string         `json:"url,omitempty"`
	Status  Status         `json:"status,omitempty"`
	Created time.Time      `json:"created"`
}

// Registry is the full set of croft-managed worktrees for one project, keyed
// by slug.
type Registry struct {
	Worktrees map[string]Worktree `json:"worktrees"`
}

// Store persists a Registry as JSON in a single directory.
type Store struct {
	dir string
}

const registryFile = "registry.json"

// Open returns a Store for the named project under the XDG data directory
// (XDG_DATA_HOME, or ~/.local/share).
func Open(project string) (*Store, error) {
	base, err := env.DataHome()
	if err != nil {
		return nil, err
	}
	return OpenAt(filepath.Join(base, "croft", project))
}

// OpenAt returns a Store rooted at an explicit directory, creating it if
// necessary. Used directly in tests.
func OpenAt(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create state dir %s: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// Dir reports the store's backing directory.
func (s *Store) Dir() string { return s.dir }

// Load reads the registry. A missing file yields an empty registry.
func (s *Store) Load() (Registry, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, registryFile))
	if os.IsNotExist(err) {
		return Registry{Worktrees: map[string]Worktree{}}, nil
	}
	if err != nil {
		return Registry{}, fmt.Errorf("read registry: %w", err)
	}
	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return Registry{}, fmt.Errorf("parse registry: %w", err)
	}
	if r.Worktrees == nil {
		r.Worktrees = map[string]Worktree{}
	}
	return r, nil
}

// Save atomically writes the registry via a temp file and rename.
func (s *Store) Save(r Registry) error {
	if r.Worktrees == nil {
		r.Worktrees = map[string]Worktree{}
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("encode registry: %w", err)
	}
	tmp, err := os.CreateTemp(s.dir, registryFile+".*")
	if err != nil {
		return fmt.Errorf("create temp registry: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp registry: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp registry: %w", err)
	}
	if err := os.Rename(tmpName, filepath.Join(s.dir, registryFile)); err != nil {
		return fmt.Errorf("commit registry: %w", err)
	}
	return nil
}

// Get returns the record for slug. found is false when no such record exists.
func (s *Store) Get(slug string) (wt Worktree, found bool, err error) {
	r, err := s.Load()
	if err != nil {
		return Worktree{}, false, err
	}
	wt, found = r.Worktrees[slug]
	return wt, found, nil
}

// Put inserts or replaces a worktree record. The read-modify-write runs under
// an exclusive registry lock so concurrent croft invocations cannot clobber
// each other.
func (s *Store) Put(wt Worktree) error {
	return s.withLock(func() error {
		r, err := s.Load()
		if err != nil {
			return err
		}
		r.Worktrees[wt.Slug] = wt
		return s.Save(r)
	})
}

// Delete removes the record for slug. Deleting an absent slug is not an error
// (idempotent teardown). The read-modify-write runs under an exclusive
// registry lock.
func (s *Store) Delete(slug string) error {
	return s.withLock(func() error {
		r, err := s.Load()
		if err != nil {
			return err
		}
		delete(r.Worktrees, slug)
		return s.Save(r)
	})
}
