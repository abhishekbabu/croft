package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoneMultiplexerIsInert(t *testing.T) {
	require.NoError(t, NoneMultiplexer{}.CreateSession("x", "/tmp", nil))
	require.False(t, NoneMultiplexer{}.Managed(), "the none multiplexer hosts nothing")
	require.False(t, NoneMultiplexer{}.HasWindow("x", "dev"))
}
