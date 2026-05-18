package sh

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCaptureStdout(t *testing.T) {
	out, err := Capture("echo", "", nil, "hello", "world")
	require.NoError(t, err)
	require.Equal(t, "hello world", strings.TrimSpace(out))
}

func TestCaptureRunsInDir(t *testing.T) {
	dir := t.TempDir()
	out, err := Capture("pwd", dir, nil)
	require.NoError(t, err)
	// macOS resolves temp dirs through /private; compare the suffix.
	require.True(t, strings.HasSuffix(strings.TrimSpace(out), dir),
		"pwd %q should run in %q", strings.TrimSpace(out), dir)
}

func TestCapturePassesEnv(t *testing.T) {
	out, err := Capture("sh", "", []string{"CROFT_TEST=on"}, "-c", "echo $CROFT_TEST")
	require.NoError(t, err)
	require.Equal(t, "on", strings.TrimSpace(out))
}

func TestCaptureErrorAnnotatesStderr(t *testing.T) {
	_, err := Capture("sh", "", nil, "-c", "echo boom >&2; exit 3")
	require.Error(t, err)
	require.Contains(t, err.Error(), "boom", "error should carry the command's stderr")
}

func TestCaptureTimeout(t *testing.T) {
	prev := DefaultTimeout
	DefaultTimeout = 50 * time.Millisecond
	t.Cleanup(func() { DefaultTimeout = prev })

	start := time.Now()
	_, err := Capture("sleep", "", nil, "5")
	require.Error(t, err)
	require.Contains(t, err.Error(), "timed out", "exceeding DefaultTimeout should be a timeout error")
	require.Less(t, time.Since(start), 2*time.Second, "the command should be killed promptly")
}

func TestCaptureCancelledBaseContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	SetBaseContext(ctx)
	t.Cleanup(func() { SetBaseContext(context.Background()) })

	_, err := Capture("sleep", "", nil, "5")
	require.Error(t, err, "a cancelled base context must abort the command")
}

func TestStreamTo(t *testing.T) {
	var buf strings.Builder
	require.NoError(t, StreamTo(&buf, "sh", "", nil, "-c", "echo out; echo err >&2"))
	require.Contains(t, buf.String(), "out")
	require.Contains(t, buf.String(), "err", "StreamTo should capture stderr too")
}

func TestLook(t *testing.T) {
	require.True(t, Look("sh"), "sh should resolve on PATH")
	require.False(t, Look("croft-definitely-not-a-real-binary"), "a bogus name should not resolve")
}
