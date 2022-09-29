package cache_test

import (
	"context"
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/cache"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	v1 "github.com/waterfall-foundation/coordinator/beacon-chain/state/v1"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/testing/assert"
	"github.com/waterfall-foundation/coordinator/testing/require"
)

func TestSkipSlotCache_RoundTrip(t *testing.T) {
	ctx := context.Background()
	c := cache.NewSkipSlotCache()

	r := [32]byte{'a'}
	s, err := c.Get(ctx, r)
	require.NoError(t, err)
	assert.Equal(t, state.BeaconState(nil), s, "Empty cache returned an object")

	require.NoError(t, c.MarkInProgress(r))

	s, err = v1.InitializeFromProto(&ethpb.BeaconState{
		Slot: 10,
	})
	require.NoError(t, err)

	c.Put(ctx, r, s)
	c.MarkNotInProgress(r)

	res, err := c.Get(ctx, r)
	require.NoError(t, err)
	assert.DeepEqual(t, res.CloneInnerState(), s.CloneInnerState(), "Expected equal protos to return from cache")
}
