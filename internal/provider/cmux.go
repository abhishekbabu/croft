package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/abhishekbabu/croft/internal/env"
	"github.com/abhishekbabu/croft/internal/sh"
)

// CmuxMultiplexer drives sessions with cmux. cmux's model maps onto the
// Multiplexer interface as: a cmux *workspace* is the session, a cmux *surface*
// (terminal) is a window.
//
// cmux materializes a surface's terminal only while it is rendered in the
// focused window — it has no detached-server model like tmux. So croft cannot
// drive a surface in a background workspace. The reliable path is to split the
// surface croft itself runs in ($CMUX_SURFACE_ID): croft focuses that surface,
// splits it (the split is live), sends the command, then moves the live
// surface into the worktree's workspace. If croft's surface cannot be focused
// — e.g. croft was not run from a cmux terminal — the operation refuses rather
// than producing a dead surface. Split surfaces are not "tabs" and cannot be
// renamed, so croft tracks window-name -> surface-id in its own state file.
type CmuxMultiplexer struct {
	bin       string
	surfaceID string // $CMUX_SURFACE_ID — croft's own surface
	stateDir  string // where the window->surface map is persisted
}

// NewCmuxMultiplexer returns a cmux-backed multiplexer. An empty bin resolves
// cmux from PATH. stateDir is where the window-tracking map is stored.
func NewCmuxMultiplexer(bin, stateDir string) *CmuxMultiplexer {
	if bin == "" {
		bin = "cmux"
	}
	return &CmuxMultiplexer{
		bin:       bin,
		surfaceID: env.CmuxSurfaceID(),
		stateDir:  stateDir,
	}
}

// cmux surface-lifecycle timing: a freshly split surface's shell needs a moment
// to accept input, and the command must register before the surface is moved
// off the focused workspace.
const (
	cmuxFocusDelay  = 250 * time.Millisecond
	cmuxReadyDelay  = 1500 * time.Millisecond
	cmuxSettleDelay = 800 * time.Millisecond
)

// Managed reports true: cmux hosts real workspaces.
func (c *CmuxMultiplexer) Managed() bool { return true }

// requireSurface returns an error when croft is not running inside cmux.
func (c *CmuxMultiplexer) requireSurface() error {
	if c.surfaceID == "" {
		return fmt.Errorf("cmux multiplexer requires running inside a cmux session ($CMUX_SURFACE_ID is unset)")
	}
	return nil
}

// ensureFocused focuses croft's own surface and verifies it took. cmux only
// gives a surface a live terminal while it is focused, so a split spawned from
// an unfocused surface is dead. This is croft's enforcement that a cmux
// operation only runs where it actually works.
func (c *CmuxMultiplexer) ensureFocused() error {
	if err := c.requireSurface(); err != nil {
		return err
	}
	// Focusing croft's own surface is legitimate — it is where the user just
	// ran croft. It also recovers focus if a previous split stole it.
	_, _ = sh.Capture(c.bin, "", nil, "focus-panel", "--panel", c.surfaceID)
	time.Sleep(cmuxFocusDelay)

	res, err := sh.Capture(c.bin, "", nil, "identify", "--json")
	if err != nil {
		return fmt.Errorf("cmux: identify: %w", err)
	}
	var id struct {
		Caller struct {
			SurfaceRef string `json:"surface_ref"`
		} `json:"caller"`
		Focused struct {
			SurfaceRef string `json:"surface_ref"`
		} `json:"focused"`
	}
	if err := json.Unmarshal([]byte(res), &id); err != nil {
		return fmt.Errorf("cmux: parse identify: %w", err)
	}
	if id.Caller.SurfaceRef == "" || id.Caller.SurfaceRef != id.Focused.SurfaceRef {
		return fmt.Errorf("cmux: croft's surface is not focused — run croft from a " +
			"focused cmux terminal (cmux only drives surfaces that are on screen)")
	}
	return nil
}

// --- cmux tree model ---

type cmuxTree struct {
	Windows []struct {
		Workspaces []cmuxWorkspace `json:"workspaces"`
	} `json:"windows"`
}

type cmuxWorkspace struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Panes []struct {
		Surfaces []struct {
			ID string `json:"id"`
		} `json:"surfaces"`
	} `json:"panes"`
}

// hasSurface reports whether the workspace contains the given surface id.
func (w cmuxWorkspace) hasSurface(id string) bool {
	for _, p := range w.Panes {
		for _, s := range p.Surfaces {
			if s.ID == id {
				return true
			}
		}
	}
	return false
}

// loadTree returns the full window/workspace/pane/surface tree.
func (c *CmuxMultiplexer) loadTree() (cmuxTree, error) {
	res, err := sh.Capture(c.bin, "", nil, "--json", "--id-format", "uuids", "tree", "--all")
	if err != nil {
		return cmuxTree{}, err
	}
	var t cmuxTree
	if err := json.Unmarshal([]byte(res), &t); err != nil {
		return cmuxTree{}, fmt.Errorf("cmux: parse tree: %w", err)
	}
	return t, nil
}

// findWorkspace resolves a workspace by its title.
func (c *CmuxMultiplexer) findWorkspace(name string) (cmuxWorkspace, bool) {
	tree, err := c.loadTree()
	if err != nil {
		return cmuxWorkspace{}, false
	}
	for _, win := range tree.Windows {
		for _, ws := range win.Workspaces {
			if ws.Title == name {
				return ws, true
			}
		}
	}
	return cmuxWorkspace{}, false
}

// CreateSession creates a cmux workspace named name, rooted at dir. Idempotent.
// env travels with each RunWindow command rather than being set on the session.
func (c *CmuxMultiplexer) CreateSession(name, dir string, _ map[string]string) error {
	if err := c.requireSurface(); err != nil {
		return err
	}
	if _, ok := c.findWorkspace(name); ok {
		return nil
	}
	_, err := sh.Capture(c.bin, "", nil, "new-workspace", "--name", name, "--cwd", dir)
	return err
}

// RunWindow launches argv as a window of the worktree's workspace. It focuses
// croft's own (live) surface, splits it, sends the command there, then moves
// the live surface into the worktree's workspace.
func (c *CmuxMultiplexer) RunWindow(name, window, _ string, env map[string]string, argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("RunWindow: empty argv")
	}
	if err := c.ensureFocused(); err != nil {
		return err
	}
	ws, ok := c.findWorkspace(name)
	if !ok {
		return fmt.Errorf("cmux: no workspace %q", name)
	}

	res, err := sh.Capture(c.bin, "", nil, "rpc", "surface.split",
		fmt.Sprintf(`{"direction":"right","surface_id":%q}`, c.surfaceID))
	if err != nil {
		return fmt.Errorf("cmux: surface.split: %w", err)
	}
	surfaceID, err := parseSurfaceSplit(res)
	if err != nil {
		return err
	}

	// Send the command while the split is still in croft's focused workspace,
	// where its terminal is live; only then move it to the worktree workspace.
	time.Sleep(cmuxReadyDelay)
	if _, err := sh.Capture(c.bin, "", nil, "send", "--surface", surfaceID, "--", cmuxCommandLine(env, argv)); err != nil {
		return fmt.Errorf("cmux: send: %w", err)
	}
	if _, err := sh.Capture(c.bin, "", nil, "send-key", "--surface", surfaceID, "enter"); err != nil {
		return fmt.Errorf("cmux: send-key: %w", err)
	}
	time.Sleep(cmuxSettleDelay)
	if _, err := sh.Capture(c.bin, "", nil, "move-surface", "--surface", surfaceID,
		"--workspace", ws.ID, "--focus", "false"); err != nil {
		return fmt.Errorf("cmux: move-surface: %w", err)
	}

	return c.trackWindow(name, window, surfaceID)
}

// HasWindow reports whether the named window's surface still exists.
func (c *CmuxMultiplexer) HasWindow(name, window string) bool {
	id, ok := c.windowSurface(name, window)
	if !ok {
		return false
	}
	tree, err := c.loadTree()
	if err != nil {
		return false
	}
	for _, win := range tree.Windows {
		for _, ws := range win.Workspaces {
			if ws.hasSurface(id) {
				return true
			}
		}
	}
	return false
}

// CapturePane returns the last n lines of the named window's surface. cmux only
// serves terminal text for a focused surface, so this is best-effort.
func (c *CmuxMultiplexer) CapturePane(name, window string, lines int) (string, error) {
	id, ok := c.windowSurface(name, window)
	if !ok {
		return "", fmt.Errorf("cmux: no window %q in workspace %q", window, name)
	}
	res, err := sh.Capture(c.bin, "", nil, "read-screen", "--surface", id, "--lines", strconv.Itoa(lines))
	return res, err
}

// Kill closes the workspace and forgets its tracked windows. Idempotent.
func (c *CmuxMultiplexer) Kill(name string) error {
	ws, ok := c.findWorkspace(name)
	if ok {
		if _, err := sh.Capture(c.bin, "", nil, "close-workspace", "--workspace", ws.ID); err != nil {
			return err
		}
	}
	return c.forgetWorkspace(name)
}

// Attach selects (focuses) the workspace in the cmux UI.
func (c *CmuxMultiplexer) Attach(name string) error {
	ws, ok := c.findWorkspace(name)
	if !ok {
		return fmt.Errorf("cmux: no workspace %q", name)
	}
	_, err := sh.Capture(c.bin, "", nil, "select-workspace", "--workspace", ws.ID)
	return err
}

// --- window -> surface tracking ---

// windowMap is workspace name -> window name -> surface id.
type windowMap map[string]map[string]string

// mapPath is the location of the persisted window-tracking map.
func (c *CmuxMultiplexer) mapPath() string {
	return filepath.Join(c.stateDir, "cmux-windows.json")
}

// loadMap reads the window-tracking map; a missing file is an empty map.
func (c *CmuxMultiplexer) loadMap() windowMap {
	data, err := os.ReadFile(c.mapPath())
	if err != nil {
		return windowMap{}
	}
	var m windowMap
	if json.Unmarshal(data, &m) != nil || m == nil {
		return windowMap{}
	}
	return m
}

// saveMap persists the window-tracking map.
func (c *CmuxMultiplexer) saveMap(m windowMap) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(c.stateDir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(c.mapPath(), data, 0o600)
}

// trackWindow records that workspace name's window maps to surfaceID.
func (c *CmuxMultiplexer) trackWindow(name, window, surfaceID string) error {
	m := c.loadMap()
	if m[name] == nil {
		m[name] = map[string]string{}
	}
	m[name][window] = surfaceID
	return c.saveMap(m)
}

// windowSurface returns the tracked surface id for a workspace's window.
func (c *CmuxMultiplexer) windowSurface(name, window string) (string, bool) {
	m := c.loadMap()
	if windows, ok := m[name]; ok {
		id, ok := windows[window]
		return id, ok
	}
	return "", false
}

// forgetWorkspace drops every tracked window for a workspace.
func (c *CmuxMultiplexer) forgetWorkspace(name string) error {
	m := c.loadMap()
	if _, ok := m[name]; !ok {
		return nil
	}
	delete(m, name)
	return c.saveMap(m)
}

// parseSurfaceSplit extracts the new surface id from an `rpc surface.split`
// response (the id appears at .surface_id or .result.surface_id).
func parseSurfaceSplit(out string) (string, error) {
	var r struct {
		SurfaceID string `json:"surface_id"`
		Result    struct {
			SurfaceID string `json:"surface_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		return "", fmt.Errorf("cmux: parse surface.split response: %w", err)
	}
	if r.SurfaceID != "" {
		return r.SurfaceID, nil
	}
	if r.Result.SurfaceID != "" {
		return r.Result.SurfaceID, nil
	}
	return "", fmt.Errorf("cmux: surface.split returned no surface id: %s", strings.TrimSpace(out))
}

// cmuxCommandLine renders argv as a shell command line with env exported via an
// `env` prefix — cmux has no per-session environment.
func cmuxCommandLine(env map[string]string, argv []string) string {
	parts := []string{"env"}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, shellQuote(k+"="+env[k]))
	}
	for _, a := range argv {
		parts = append(parts, shellQuote(a))
	}
	return strings.Join(parts, " ")
}

// shellQuote single-quotes a string for safe inclusion in a shell command.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
