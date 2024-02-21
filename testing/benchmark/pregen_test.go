package benchmark

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func TestPreGenFullBlock(t *testing.T) {
	t.Skip() // Generate test data
	_, err := PreGenFullBlock()
	require.NoError(t, err)
}

func TestPreGenState1Epoch(t *testing.T) {
	t.Skip() // Generate test data
	_, err := PreGenFullBlock()
	require.NoError(t, err)
}

func TestPreGenstateFullEpochs(t *testing.T) {
	t.Skip() // Generate test data
	_, err := PreGenFullBlock()
	require.NoError(t, err)
}
