package provider

import "errors"

// Coordination spawns and coordinates named agents across worktrees.
type Coordination interface {
	// Spawn launches a coordinated agent.
	Spawn(spec PeerSpec) (Peer, error)
	// Status lists known peers.
	Status() ([]Peer, error)
	// Dispatch sends a message to a peer.
	Dispatch(target Peer, msg string) error
}

// NoopCoordination is the inert coordination provider, used when no
// coordination backend is selected. It rejects spawn/dispatch and reports no
// peers, so core never crashes on the coordination slot.
type NoopCoordination struct{}

// errCoordinationUnavailable is returned by the inert provider.
var errCoordinationUnavailable = errors.New("no coordination provider configured")

// Spawn reports that coordination is unavailable.
func (NoopCoordination) Spawn(PeerSpec) (Peer, error) { return Peer{}, errCoordinationUnavailable }

// Status returns no peers.
func (NoopCoordination) Status() ([]Peer, error) { return nil, nil }

// Dispatch reports that coordination is unavailable.
func (NoopCoordination) Dispatch(Peer, string) error { return errCoordinationUnavailable }

// AgentLauncher builds the process argv and environment for the agent named by
// a PeerSpec. It lets the coordination provider spawn agents without the
// provider package depending on the agent package.
type AgentLauncher func(spec PeerSpec) (argv []string, env map[string]string, err error)

// PeerSpec describes a coordinated agent to spawn (used by Coordination).
type PeerSpec struct {
	Name  string
	Agent string
	Role  string
	Dir   string
}

// PeerStatus is a coordinated peer's lifecycle state.
type PeerStatus string

// PeerStatusWorking marks a peer whose agent has been launched.
const PeerStatusWorking PeerStatus = "working"

// Peer is a spawned, coordinated agent.
type Peer struct {
	Name   string
	Agent  string
	Status PeerStatus
}
