package state

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenUsesXDGDataHome(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	s, err := Open("demo")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dataHome, "croft", "demo"), s.Dir())

	r, err := s.Load()
	require.NoError(t, err)
	require.Empty(t, r.Worktrees, "a fresh store has an empty registry")
}

func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := openAt(t.TempDir())
	require.NoError(t, err, "openAt")
	return s
}

func TestLoadEmptyRegistry(t *testing.T) {
	r, err := newStore(t).Load()
	require.NoError(t, err)
	require.NotNil(t, r.Worktrees)
	require.Empty(t, r.Worktrees)
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
	require.NoError(t, s.Put(wt))

	got, found, err := s.Get("feat")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "my-feature", got.Branch)
	require.Equal(t, 3000, got.Ports["api"])
	require.Equal(t, StatusWorking, got.Status)

	require.NoError(t, s.Delete("feat"))
	_, found, _ = s.Get("feat")
	require.False(t, found, "record should be gone after Delete")

	// Deleting an absent slug must be a no-op, not an error.
	require.NoError(t, s.Delete("feat"), "Delete of an absent slug should be idempotent")
}

func TestSaveLoadRoundTrip(t *testing.T) {
	s := newStore(t)
	in := Registry{Worktrees: map[string]Worktree{
		"a": {Slug: "a", Branch: "ba"},
		"b": {Slug: "b", Branch: "bb"},
	}}
	require.NoError(t, s.Save(in))

	out, err := s.Load()
	require.NoError(t, err)
	require.Len(t, out.Worktrees, 2)
	require.Equal(t, "bb", out.Worktrees["b"].Branch)

	// Reopening the same directory must see persisted state.
	s2, err := openAt(s.Dir())
	require.NoError(t, err)
	r, err := s2.Load()
	require.NoError(t, err)
	require.Len(t, r.Worktrees, 2, "state did not persist across Store instances")
}
