package cache_test

import (
	"context"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"google.golang.org/protobuf/proto"
)

func TestAttestationCache_RoundTrip(t *testing.T) {
	ctx := context.Background()
	c := cache.NewAttestationCache()

	req := &ethpb.AttestationDataRequest{
		CommitteeIndex: 0,
		Slot:           1,
	}

	response, err := c.Get(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, (*ethpb.AttestationData)(nil), response)

	assert.NoError(t, c.MarkInProgress(req))

	res := &ethpb.AttestationData{
		Target: &ethpb.Checkpoint{Epoch: 5, Root: make([]byte, 32)},
	}

	assert.NoError(t, c.Put(ctx, req, res))
	assert.NoError(t, c.MarkNotInProgress(req))

	response, err = c.Get(ctx, req)
	assert.NoError(t, err)

	if !proto.Equal(response, res) {
		t.Error("Expected equal protos to return from cache")
	}
}
