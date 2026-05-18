package worktree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/abhishekbabu/croft/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestParseWorktreeList(t *testing.T) {
	out := "worktree /repo\n" +
		"HEAD abc123\n" +
		"branch refs/heads/main\n" +
		"\n" +
		"worktree /repo/../wt/demo.feat\n" +
		"HEAD def456\n" +
		"detached\n"
	got := parseWorktreeList(out)
	require.Len(t, got, 2)
	require.Equal(t, "main", got[0].Branch)
	require.Equal(t, "/repo", got[0].Path)
	require.True(t, got[1].Detached, "second entry should be detached")
	require.Empty(t, got[1].Branch)
}

func TestManagerAddListRemove(t *testing.T) {
	repo := testutil.GitRepo(t)
	mgr := NewManager(repo)
	wtPath := filepath.Join(t.TempDir(), "demo.feat")

	require.False(t, mgr.BranchExists("feat"), "feat branch should not exist yet")
	require.NoError(t, mgr.Add(wtPath, "feat", ""))
	require.True(t, mgr.BranchExists("feat"), "feat branch should exist after Add")
	require.DirExists(t, wtPath)

	list, err := mgr.List()
	require.NoError(t, err)
	var found bool
	for _, w := range list {
		if w.Branch == "feat" {
			found = true
		}
	}
	require.True(t, found, "List did not include the feat worktree: %+v", list)

	require.NoError(t, mgr.Remove(wtPath, false))
	require.NoDirExists(t, wtPath, "worktree directory should be gone after Remove")
	require.NoError(t, mgr.Prune())
}

func TestDirtyAndStash(t *testing.T) {
	repo := testutil.GitRepo(t)
	mgr := NewManager(repo)

	require.False(t, mgr.IsDirty(repo), "fresh repo should be clean")
	stashed, err := mgr.Stash(repo, "test")
	require.NoError(t, err)
	require.False(t, stashed, "nothing to stash in a clean tree")

	require.NoError(t, os.WriteFile(filepath.Join(repo, "f"), []byte("changed"), 0o644))
	require.True(t, mgr.IsDirty(repo), "repo should be dirty after edit")

	stashed, err = mgr.Stash(repo, "test")
	require.NoError(t, err)
	require.True(t, stashed, "dirty tree should be stashed")
	require.False(t, mgr.IsDirty(repo), "repo should be clean after stash")

	require.NoError(t, mgr.StashPop(repo))
	require.True(t, mgr.IsDirty(repo), "changes should be restored after StashPop")
}

func TestInRebase(t *testing.T) {
	repo := testutil.GitRepo(t)
	mgr := NewManager(repo)
	require.False(t, mgr.InRebase(repo), "fresh repo should not be mid-rebase")

	// Simulate a stalled rebase by creating the state directory git uses.
	gd, err := mgr.GitDir(repo)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(gd, "rebase-merge"), 0o755))
	require.True(t, mgr.InRebase(repo), "InRebase should detect the rebase-merge directory")
}
