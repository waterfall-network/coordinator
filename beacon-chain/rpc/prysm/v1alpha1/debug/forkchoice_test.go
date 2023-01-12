package debug

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	mock "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/forkchoice/protoarray"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func TestServer_GetForkChoice_ProtoArray(t *testing.T) {
	store := protoarray.New(0, 0, [32]byte{'a'})
	bs := &Server{HeadFetcher: &mock.ChainService{ForkChoiceStore: store}}
	res, err := bs.GetForkChoice(context.Background(), &empty.Empty{})
	require.NoError(t, err)
	assert.Equal(t, store.JustifiedEpoch(), res.JustifiedEpoch, "Did not get wanted justified epoch")
	assert.Equal(t, store.FinalizedEpoch(), res.FinalizedEpoch, "Did not get wanted finalized epoch")
}
