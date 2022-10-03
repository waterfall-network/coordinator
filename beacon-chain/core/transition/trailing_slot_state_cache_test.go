package transition_test

import (
	"context"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/transition"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/util"
)

func TestTrailingSlotState_RoundTrip(t *testing.T) {
	ctx := context.Background()
	r := []byte{'a'}
	s, err := transition.NextSlotState(ctx, r)
	require.NoError(t, err)
	require.Equal(t, nil, s)

	s, _ = util.DeterministicGenesisState(t, 1)
	require.NoError(t, transition.UpdateNextSlotCache(ctx, r, s))
	s, err = transition.NextSlotState(ctx, r)
	require.NoError(t, err)
	require.Equal(t, types.Slot(1), s.Slot())

	require.NoError(t, transition.UpdateNextSlotCache(ctx, r, s))
	s, err = transition.NextSlotState(ctx, r)
	require.NoError(t, err)
	require.Equal(t, types.Slot(2), s.Slot())
}
