package worktree

import "testing"

func TestAllocatePortsLowestFree(t *testing.T) {
	got, err := AllocatePorts(3000, 3999, []string{"api", "db"}, nil)
	if err != nil {
		t.Fatalf("AllocatePorts: %v", err)
	}
	if got["api"] != 3000 || got["db"] != 3001 {
		t.Errorf("AllocatePorts = %v, want api=3000 db=3001", got)
	}
}

func TestAllocatePortsSkipsTaken(t *testing.T) {
	taken := map[int]bool{3000: true, 3001: true}
	got, err := AllocatePorts(3000, 3999, []string{"api"}, taken)
	if err != nil {
		t.Fatalf("AllocatePorts: %v", err)
	}
	if got["api"] != 3002 {
		t.Errorf("AllocatePorts skipping taken = %v, want api=3002", got)
	}
	if taken[3002] {
		t.Error("AllocatePorts must not mutate the taken set")
	}
}

func TestAllocatePortsExhausted(t *testing.T) {
	if _, err := AllocatePorts(3000, 3001, []string{"a", "b", "c"}, nil); err == nil {
		t.Error("expected exhaustion error when more services than ports")
	}
}
