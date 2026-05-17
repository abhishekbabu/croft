package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmdRuns(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("root command failed: %v", err)
	}
}

func TestRootCmdVersion(t *testing.T) {
	cmd := newRootCmd()
	out := new(bytes.Buffer)
	cmd.SetArgs([]string{"--version"})
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("--version failed: %v", err)
	}
	if !strings.Contains(out.String(), "croft") {
		t.Fatalf("version output missing tool name: %q", out.String())
	}
}
