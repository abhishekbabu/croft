package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodexLaunch(t *testing.T) {
	r := &CodexRunner{bin: "codex"}

	inv, err := r.Launch(Spec{Headless: true, Profile: "dev", Prompt: "do it"})
	require.NoError(t, err)
	require.Subset(t, inv.Args, []string{"exec", "--json", "--profile", "dev"})

	interactive, err := r.Launch(Spec{})
	require.NoError(t, err)
	require.NotContains(t, interactive.Args, "exec", "interactive codex launch should not use exec")

	resume, err := r.Resume("th-1", Spec{Headless: true})
	require.NoError(t, err)
	require.Contains(t, resume.Args, "resume")
	require.Contains(t, resume.Args, "th-1")
}
