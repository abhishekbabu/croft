package state

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// lockFile is the advisory-lock file guarding registry mutations. It is
// separate from registry.json so the atomic temp-file rename in Save never
// disturbs the lock's inode.
const lockFile = "registry.lock"

// withLock runs fn while holding an exclusive advisory lock on the registry.
// croft mutations are read-modify-write (Load, change, Save); without the lock
// two concurrent croft invocations could each Load the same registry and the
// second Save would silently clobber the first's change. flock serializes the
// whole cycle across processes.
func (s *Store) withLock(fn func() error) error {
	f, err := os.OpenFile(filepath.Join(s.dir, lockFile), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("open registry lock: %w", err)
	}
	defer func() { _ = f.Close() }()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("lock registry: %w", err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	return fn()
}
