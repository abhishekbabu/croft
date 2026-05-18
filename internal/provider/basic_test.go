package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// stubMux is an inert Multiplexer: every call succeeds and does nothing, so
// BasicCoordination tests exercise only the coordination logic.
type stubMux struct{}

func (stubMux) Managed() bool                                         { return false }
func (stubMux) CreateSession(string, string, map[string]string) error { return nil }
func (stubMux) RunWindow(string, string, string, map[string]string, []string) error {
	return nil
}
func (stubMux) HasWindow(string, string) bool { return false }
func (stubMux) Attach(string) error           { return nil }
func (stubMux) Kill(string) error             { return nil }
func (stubMux) CapturePane(string, string, int) (string, error) {
	return "", nil
}

// stubLauncher is an AgentLauncher that returns a fixed, harmless invocation.
func stubLauncher(PeerSpec) ([]string, map[string]string, error) {
	return []string{"true"}, map[string]string{"K": "v"}, nil
}

func TestBasicCoordinationSpawnStatusDispatch(t *testing.T) {
	dir := t.TempDir()
	c := NewBasicCoordination(dir, "sess", stubMux{}, stubLauncher)

	peers, err := c.Status()
	require.NoError(t, err)
	require.Empty(t, peers, "a fresh coordination dir has no peers")

	p, err := c.Spawn(PeerSpec{Name: "worker", Agent: "claude", Role: "reviewer", Dir: dir})
	require.NoError(t, err)
	require.Equal(t, "worker", p.Name)
	require.Equal(t, "claude", p.Agent)
	require.Equal(t, PeerStatusWorking, p.Status)

	// Status reflects the spawned peer — and a fresh instance reads the same
	// on-disk record, exercising savePeer/loadPeer.
	peers, err = NewBasicCoordination(dir, "sess", stubMux{}, stubLauncher).Status()
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, "worker", peers[0].Name)
	require.Equal(t, PeerStatusWorking, peers[0].Status)

	// Dispatch appends to the peer's mailbox.
	require.NoError(t, c.Dispatch(Peer{Name: "worker"}, "ship it"))
	inbox, err := os.ReadFile(filepath.Join(dir, "worker.inbox"))
	require.NoError(t, err)
	require.Contains(t, string(inbox), "ship it")
}

func TestBasicCoordinationRejects(t *testing.T) {
	c := NewBasicCoordination(t.TempDir(), "sess", stubMux{}, stubLauncher)

	_, err := c.Spawn(PeerSpec{Name: "  "})
	require.Error(t, err, "a blank peer name should be rejected")

	require.Error(t, c.Dispatch(Peer{Name: "ghost"}, "hello"),
		"dispatch to an unknown peer should fail")
}
