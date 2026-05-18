package provider

import (
	"fmt"
	"os"
	"strings"

	"github.com/abhishekbabu/croft/internal/sh"
)

// ComposeInfra runs a per-worktree `docker compose` stack. Each worktree gets
// its own compose project namespace (croft-<slug>) so stacks never collide.
type ComposeInfra struct {
	bin string
}

// NewComposeInfra returns a docker-compose-backed infra provider. An empty bin
// resolves docker from PATH.
func NewComposeInfra(bin string) *ComposeInfra {
	if bin == "" {
		bin = "docker"
	}
	return &ComposeInfra{bin: bin}
}

// compose runs `docker compose -p <project> <args...>` in the worktree, with
// croft's environment exported so the compose file can read ${SERVICE_PORT}.
func (c *ComposeInfra) compose(wt Worktree, args ...string) (string, error) {
	full := append([]string{"compose", "-p", ProjectName(wt)}, args...)
	env := append(os.Environ(), envSlice(Env(wt))...)
	return sh.Capture(c.bin, wt.Path, env, full...)
}

// Up brings the worktree's compose stack online in the background.
func (c *ComposeInfra) Up(wt Worktree) error {
	_, err := c.compose(wt, "up", "-d")
	return err
}

// Down tears the worktree's compose stack down, removing volumes.
func (c *ComposeInfra) Down(wt Worktree) error {
	_, err := c.compose(wt, "down", "-v")
	return err
}

// Status reports the stack as up when the compose project has containers.
func (c *ComposeInfra) Status(wt Worktree) (InfraState, error) {
	res, err := c.compose(wt, "ps", "--quiet")
	if err != nil {
		return InfraState{}, err
	}
	ids := strings.Fields(res)
	if len(ids) == 0 {
		return InfraState{Up: false, Detail: "no containers"}, nil
	}
	noun := "containers"
	if len(ids) == 1 {
		noun = "container"
	}
	return InfraState{Up: true, Detail: fmt.Sprintf("%d %s", len(ids), noun)}, nil
}
