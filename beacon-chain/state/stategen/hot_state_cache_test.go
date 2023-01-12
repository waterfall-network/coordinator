package stategen

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v1"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func TestHotStateCache_RoundTrip(t *testing.T) {
	c := newHotStateCache()
	root := [32]byte{'A'}
	s := c.get(root)
	assert.Equal(t, state.BeaconState(nil), s)
	assert.Equal(t, false, c.has(root), "Empty cache has an object")

	s, err := v1.InitializeFromProto(&ethpb.BeaconState{
		Slot: 10,
	})
	require.NoError(t, err)

	c.put(root, s)
	assert.Equal(t, true, c.has(root), "Empty cache does not have an object")

	res := c.get(root)
	assert.NotNil(t, s)
	assert.DeepEqual(t, res.CloneInnerState(), s.CloneInnerState(), "Expected equal protos to return from cache")

	c.delete(root)
	assert.Equal(t, false, c.has(root), "Cache not supposed to have the object")
}
