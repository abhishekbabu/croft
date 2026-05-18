package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// peerRecord is the on-disk representation of a spawned peer.
type peerRecord struct {
	Name    string     `json:"name"`
	Agent   string     `json:"agent"`
	Role    string     `json:"role,omitempty"`
	Dir     string     `json:"dir"`
	Status  PeerStatus `json:"status"`
	Created time.Time  `json:"created"`
}

// BasicCoordination is the agent-agnostic coordination backend: file-based
// peer registry, multiplexer-hosted agents, and a file mailbox for dispatch.
// It works for any agent runner (PLAN.md §4, decision 3).
type BasicCoordination struct {
	dir     string // peer state directory (shared by every worktree of a project)
	session string // multiplexer session peers run in
	mux     Multiplexer
	launch  AgentLauncher
}

// NewBasicCoordination returns a file-based coordination backend.
func NewBasicCoordination(dir, session string, mux Multiplexer, launch AgentLauncher) *BasicCoordination {
	return &BasicCoordination{dir: dir, session: session, mux: mux, launch: launch}
}

// Spawn launches the peer's agent into the shared session and records it.
func (c *BasicCoordination) Spawn(spec PeerSpec) (Peer, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return Peer{}, errors.New("peer name is required")
	}
	if err := os.MkdirAll(c.dir, 0o700); err != nil {
		return Peer{}, fmt.Errorf("create peer state dir: %w", err)
	}
	argv, env, err := c.launch(spec)
	if err != nil {
		return Peer{}, err
	}
	if err := c.mux.CreateSession(c.session, spec.Dir, env); err != nil {
		return Peer{}, fmt.Errorf("create peer session: %w", err)
	}
	if err := c.mux.RunWindow(c.session, spec.Name, spec.Dir, env, argv); err != nil {
		return Peer{}, fmt.Errorf("launch peer: %w", err)
	}
	rec := peerRecord{
		Name:    spec.Name,
		Agent:   spec.Agent,
		Role:    spec.Role,
		Dir:     spec.Dir,
		Status:  PeerStatusWorking,
		Created: time.Now().UTC(),
	}
	if err := c.savePeer(rec); err != nil {
		return Peer{}, err
	}
	return Peer{Name: rec.Name, Agent: rec.Agent, Status: rec.Status}, nil
}

// Status lists every recorded peer, sorted by name.
func (c *BasicCoordination) Status() ([]Peer, error) {
	entries, err := os.ReadDir(c.dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read peer state dir: %w", err)
	}
	var peers []Peer
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		rec, err := c.loadPeer(strings.TrimSuffix(e.Name(), ".json"))
		if err != nil {
			continue
		}
		peers = append(peers, Peer{Name: rec.Name, Agent: rec.Agent, Status: rec.Status})
	}
	sort.Slice(peers, func(i, j int) bool { return peers[i].Name < peers[j].Name })
	return peers, nil
}

// Dispatch appends a message to the target peer's mailbox file.
func (c *BasicCoordination) Dispatch(target Peer, msg string) error {
	if _, err := c.loadPeer(target.Name); err != nil {
		return fmt.Errorf("no such peer %q", target.Name)
	}
	mailbox := filepath.Join(c.dir, target.Name+".inbox")
	f, err := os.OpenFile(mailbox, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open mailbox: %w", err)
	}
	defer func() { _ = f.Close() }()
	stamp := time.Now().UTC().Format(time.RFC3339)
	if _, err := fmt.Fprintf(f, "[%s] %s\n", stamp, msg); err != nil {
		return fmt.Errorf("write mailbox: %w", err)
	}
	return nil
}

// savePeer atomically writes a peer record.
func (c *BasicCoordination) savePeer(rec peerRecord) error {
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(c.dir, rec.Name+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write peer record: %w", err)
	}
	return os.Rename(tmp, path)
}

// loadPeer reads a peer record by name.
func (c *BasicCoordination) loadPeer(name string) (peerRecord, error) {
	data, err := os.ReadFile(filepath.Join(c.dir, name+".json"))
	if err != nil {
		return peerRecord{}, err
	}
	var rec peerRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return peerRecord{}, err
	}
	return rec, nil
}
