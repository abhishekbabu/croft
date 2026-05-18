package agent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecRunnerSubstitutes(t *testing.T) {
	r := &ExecRunner{name: "gemini", argv: []string{"gemini", "--cwd", "{dir}", "{prompt}"}}
	inv, err := r.Launch(Spec{Dir: "/wt/demo", Prompt: "fix bug"})
	require.NoError(t, err)
	require.Equal(t, "gemini", inv.Path)
	joined := strings.Join(inv.Args, " ")
	require.Contains(t, joined, "/wt/demo")
	require.Contains(t, joined, "fix bug")
}
