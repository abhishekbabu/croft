package provider

import (
	"sort"
	"strconv"
	"strings"

	"github.com/abhishekbabu/croft/internal/sh"
)

// Router gives a worktree a stable URL / route.
type Router interface {
	// Register routes traffic to the worktree and returns its URL.
	Register(wt Worktree) (url string, err error)
	// Release removes the worktree's route.
	Release(wt Worktree) error
}

// NoneRouter is the no-op router: worktrees get no managed route.
type NoneRouter struct{}

// Register returns an empty URL.
func (NoneRouter) Register(Worktree) (string, error) { return "", nil }

// Release does nothing.
func (NoneRouter) Release(Worktree) error { return nil }

// PortlessRouter registers a worktree's services as stable .localhost URLs via
// the portless proxy.
type PortlessRouter struct {
	bin string
}

// NewPortlessRouter returns a portless-backed router. An empty bin resolves
// portless from PATH.
func NewPortlessRouter(bin string) *PortlessRouter {
	if bin == "" {
		bin = "portless"
	}
	return &PortlessRouter{bin: bin}
}

// aliasName is the portless route name for one worktree service.
func aliasName(wt Worktree, service string) string {
	return wt.Slug + "-" + service
}

// sortedServices returns a worktree's service names in stable order.
func sortedServices(ports map[string]int) []string {
	svcs := make([]string, 0, len(ports))
	for s := range ports {
		svcs = append(svcs, s)
	}
	sort.Strings(svcs)
	return svcs
}

// Register adds a static portless route per worktree service and returns the
// URL of the first one as the worktree's URL.
func (r *PortlessRouter) Register(wt Worktree) (string, error) {
	services := sortedServices(wt.Ports)
	if len(services) == 0 {
		return "", nil
	}
	for _, svc := range services {
		name := aliasName(wt, svc)
		if _, err := sh.Capture(r.bin, "", nil, "alias", name, strconv.Itoa(wt.Ports[svc])); err != nil {
			return "", err
		}
	}
	res, err := sh.Capture(r.bin, "", nil, "get", aliasName(wt, services[0]))
	if err != nil {
		// Routes are registered; the URL lookup is best-effort.
		return "", nil
	}
	return strings.TrimSpace(res), nil
}

// Release removes the worktree's portless routes. It is best-effort: a route
// that is already gone does not block teardown.
func (r *PortlessRouter) Release(wt Worktree) error {
	for _, svc := range sortedServices(wt.Ports) {
		_, _ = sh.Capture(r.bin, "", nil, "alias", "--remove", aliasName(wt, svc))
	}
	return nil
}
