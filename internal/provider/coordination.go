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

// NoopCoordination is a placeholder coordination provider used until the
// `basic` backend lands (PLAN.md M7). It rejects spawn/dispatch and reports no
// peers, so selecting coordination never crashes core.
type NoopCoordination struct{}

// errCoordinationUnimplemented is returned by the placeholder provider.
var errCoordinationUnimplemented = errors.New("coordination provider not available yet")

// Spawn reports that coordination is not yet implemented.
func (NoopCoordination) Spawn(PeerSpec) (Peer, error) {
	return Peer{}, errCoordinationUnimplemented
}

// Status returns no peers.
func (NoopCoordination) Status() ([]Peer, error) { return nil, nil }

// Dispatch reports that coordination is not yet implemented.
func (NoopCoordination) Dispatch(Peer, string) error { return errCoordinationUnimplemented }
