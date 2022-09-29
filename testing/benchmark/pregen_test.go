package benchmark

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/require"
)

func TestPreGenFullBlock(t *testing.T) {
	_, err := PreGenFullBlock()
	require.NoError(t, err)
}

func TestPreGenState1Epoch(t *testing.T) {
	_, err := PreGenFullBlock()
	require.NoError(t, err)
}

func TestPreGenstateFullEpochs(t *testing.T) {
	_, err := PreGenFullBlock()
	require.NoError(t, err)
}
