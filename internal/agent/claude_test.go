package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClaudeLaunch(t *testing.T) {
	r := &ClaudeRunner{bin: "claude"}

	inv, err := r.Launch(Spec{Dir: "/wt", SessionID: "sid-1"})
	require.NoError(t, err)
	require.Equal(t, "claude", inv.Path)
	require.Equal(t, "/wt", inv.Dir)
	require.Contains(t, inv.Args, "--session-id")
	require.Contains(t, inv.Args, "sid-1")
	require.NotContains(t, inv.Args, "-p", "interactive launch should not be headless")

	headless, err := r.Launch(Spec{Headless: true})
	require.NoError(t, err)
	require.Contains(t, headless.Args, "-p")
	require.Contains(t, headless.Args, "--bare")
}

func TestClaudeResume(t *testing.T) {
	r := &ClaudeRunner{bin: "claude"}
	inv, err := r.Resume("sid-9", Spec{})
	require.NoError(t, err)
	require.Contains(t, inv.Args, "--resume")
	require.Contains(t, inv.Args, "sid-9")

	_, err = r.Resume("", Spec{})
	require.Error(t, err, "resume with an empty id should fail")
}
