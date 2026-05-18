package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	env := Env(Worktree{
		Slug:  "feat",
		Path:  "/wt/demo.feat",
		Ports: map[string]int{"api": 3000, "postgres": 3001},
	})
	require.Equal(t, "feat", env["CROFT_SLUG"])
	require.Equal(t, "croft-feat", env["COMPOSE_PROJECT_NAME"])
	require.Equal(t, "3000", env["API_PORT"])
	require.Equal(t, "3001", env["POSTGRES_PORT"])
}
