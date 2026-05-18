package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoneMultiplexerIsInert(t *testing.T) {
	require.NoError(t, NoneMultiplexer{}.CreateSession("x", "/tmp", nil))
	require.False(t, NoneMultiplexer{}.Managed(), "the none multiplexer hosts nothing")
	require.False(t, NoneMultiplexer{}.HasWindow("x", "dev"))

	require.Error(t, NoneMultiplexer{}.RunWindow("", "", t.TempDir(), nil, nil),
		"RunWindow with empty argv should fail")
	require.NoError(t, NoneMultiplexer{}.RunWindow("", "", t.TempDir(), nil, []string{"true"}),
		"RunWindow runs argv in the foreground")

	require.NoError(t, NoneMultiplexer{}.Attach("x"))
	require.NoError(t, NoneMultiplexer{}.Kill("x"))

	out, err := NoneMultiplexer{}.CapturePane("x", "dev", 10)
	require.NoError(t, err)
	require.Empty(t, out)
}
