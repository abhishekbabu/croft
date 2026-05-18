package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootCmdRuns(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	require.NoError(t, cmd.Execute())
}

func TestRootCmdVersion(t *testing.T) {
	cmd := newRootCmd()
	out := new(bytes.Buffer)
	cmd.SetArgs([]string{"--version"})
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	require.NoError(t, cmd.Execute())
	require.Contains(t, out.String(), "croft")
}
