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

func TestExecRunnerEmptyCommand(t *testing.T) {
	_, err := (&ExecRunner{name: "x"}).Launch(Spec{})
	require.Error(t, err, "an exec runner with no argv should fail to launch")
}

func TestExecRunnerNameAndResume(t *testing.T) {
	r := &ExecRunner{name: "gemini", argv: []string{"gemini", "{prompt}"}}
	require.Equal(t, "gemini", r.Name())

	// A generic CLI agent has no resume protocol — Resume is a fresh launch.
	inv, err := r.Resume("ignored-session", Spec{Prompt: "again"})
	require.NoError(t, err)
	require.Equal(t, []string{"gemini", "again"}, inv.Argv())
}
