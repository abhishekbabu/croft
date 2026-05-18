package state

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestConcurrentPut exercises the registry lock: many goroutines each insert a
// distinct record at once. Without locking, racing read-modify-write cycles
// would drop records; with it, every record must survive.
func TestConcurrentPut(t *testing.T) {
	s := newStore(t)
	const n = 24
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			slug := fmt.Sprintf("wt%02d", i)
			require.NoError(t, s.Put(Worktree{Slug: slug, Branch: slug}))
		}(i)
	}
	wg.Wait()

	r, err := s.Load()
	require.NoError(t, err)
	require.Len(t, r.Worktrees, n, "concurrent Puts dropped records — registry lock failed")
}
