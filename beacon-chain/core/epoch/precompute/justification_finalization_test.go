package precompute_test

import (
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/go-bitfield"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/epoch/precompute"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v1"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func TestProcessJustificationAndFinalizationPreCompute_ConsecutiveEpochs(t *testing.T) {
	e := params.BeaconConfig().FarFutureEpoch
	a := params.BeaconConfig().MaxEffectiveBalance
	blockRoots := make([][]byte, params.BeaconConfig().SlotsPerEpoch*2+1)
	for i := 0; i < len(blockRoots); i++ {
		blockRoots[i] = []byte{byte(i)}
	}
	base := &ethpb.BeaconState{
		Slot: params.BeaconConfig().SlotsPerEpoch*2 + 1,
		PreviousJustifiedCheckpoint: &ethpb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		CurrentJustifiedCheckpoint: &ethpb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		FinalizedCheckpoint: &ethpb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		JustificationBits:   bitfield.Bitvector4{0x0F}, // 0b1111
		Validators:          []*ethpb.Validator{{ExitEpoch: e}, {ExitEpoch: e}, {ExitEpoch: e}, {ExitEpoch: e}},
		Balances:            []uint64{a, a, a, a}, // validator total balance should be 128000000000
		BlockRoots:          blockRoots,
	}
	state, err := v1.InitializeFromProto(base)
	require.NoError(t, err)
	attestedBalance := 4 * uint64(e) * 3 / 2
	b := &precompute.Balance{PrevEpochTargetAttested: attestedBalance}
	newState, err := precompute.ProcessJustificationAndFinalizationPreCompute(state, b)
	require.NoError(t, err)
	rt := [32]byte{byte(64)}
	assert.DeepEqual(t, rt[:], newState.CurrentJustifiedCheckpoint().Root, "Unexpected justified root")
	assert.Equal(t, types.Epoch(2), newState.CurrentJustifiedCheckpoint().Epoch, "Unexpected justified epoch")
	assert.Equal(t, types.Epoch(0), newState.PreviousJustifiedCheckpoint().Epoch, "Unexpected previous justified epoch")
	assert.DeepEqual(t, params.BeaconConfig().ZeroHash[:], newState.FinalizedCheckpoint().Root, "Unexpected finalized root")
	assert.Equal(t, types.Epoch(0), newState.FinalizedCheckpointEpoch(), "Unexpected finalized epoch")
}

func TestProcessJustificationAndFinalizationPreCompute_JustifyCurrentEpoch(t *testing.T) {
	e := params.BeaconConfig().FarFutureEpoch
	a := params.BeaconConfig().MaxEffectiveBalance
	blockRoots := make([][]byte, params.BeaconConfig().SlotsPerEpoch*2+1)
	for i := 0; i < len(blockRoots); i++ {
		blockRoots[i] = []byte{byte(i)}
	}
	base := &ethpb.BeaconState{
		Slot: params.BeaconConfig().SlotsPerEpoch*2 + 1,
		PreviousJustifiedCheckpoint: &ethpb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		CurrentJustifiedCheckpoint: &ethpb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		FinalizedCheckpoint: &ethpb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		JustificationBits:   bitfield.Bitvector4{0x03}, // 0b0011
		Validators:          []*ethpb.Validator{{ExitEpoch: e}, {ExitEpoch: e}, {ExitEpoch: e}, {ExitEpoch: e}},
		Balances:            []uint64{a, a, a, a}, // validator total balance should be 128000000000
		BlockRoots:          blockRoots,
	}
	state, err := v1.InitializeFromProto(base)
	require.NoError(t, err)
	attestedBalance := 4 * uint64(e) * 3 / 2
	b := &precompute.Balance{PrevEpochTargetAttested: attestedBalance}
	newState, err := precompute.ProcessJustificationAndFinalizationPreCompute(state, b)
	require.NoError(t, err)
	rt := [32]byte{byte(64)}
	assert.DeepEqual(t, rt[:], newState.CurrentJustifiedCheckpoint().Root, "Unexpected current justified root")
	assert.Equal(t, types.Epoch(2), newState.CurrentJustifiedCheckpoint().Epoch, "Unexpected justified epoch")
	assert.Equal(t, types.Epoch(0), newState.PreviousJustifiedCheckpoint().Epoch, "Unexpected previous justified epoch")
	assert.DeepEqual(t, params.BeaconConfig().ZeroHash[:], newState.FinalizedCheckpoint().Root)
	assert.Equal(t, types.Epoch(0), newState.FinalizedCheckpointEpoch(), "Unexpected finalized epoch")
}

func TestProcessJustificationAndFinalizationPreCompute_JustifyPrevEpoch(t *testing.T) {
	e := params.BeaconConfig().FarFutureEpoch
	a := params.BeaconConfig().MaxEffectiveBalance
	blockRoots := make([][]byte, params.BeaconConfig().SlotsPerEpoch*2+1)
	for i := 0; i < len(blockRoots); i++ {
		blockRoots[i] = []byte{byte(i)}
	}
	base := &ethpb.BeaconState{
		Slot: params.BeaconConfig().SlotsPerEpoch*2 + 1,
		PreviousJustifiedCheckpoint: &ethpb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		CurrentJustifiedCheckpoint: &ethpb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		JustificationBits: bitfield.Bitvector4{0x03}, // 0b0011
		Validators:        []*ethpb.Validator{{ExitEpoch: e}, {ExitEpoch: e}, {ExitEpoch: e}, {ExitEpoch: e}},
		Balances:          []uint64{a, a, a, a}, // validator total balance should be 128000000000
		BlockRoots:        blockRoots, FinalizedCheckpoint: &ethpb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
	}
	state, err := v1.InitializeFromProto(base)
	require.NoError(t, err)
	attestedBalance := 4 * uint64(e) * 3 / 2
	b := &precompute.Balance{PrevEpochTargetAttested: attestedBalance}
	newState, err := precompute.ProcessJustificationAndFinalizationPreCompute(state, b)
	require.NoError(t, err)
	rt := [32]byte{byte(64)}
	assert.DeepEqual(t, rt[:], newState.CurrentJustifiedCheckpoint().Root, "Unexpected current justified root")
	assert.Equal(t, types.Epoch(0), newState.PreviousJustifiedCheckpoint().Epoch, "Unexpected previous justified epoch")
	assert.Equal(t, types.Epoch(2), newState.CurrentJustifiedCheckpoint().Epoch, "Unexpected justified epoch")
	assert.DeepEqual(t, params.BeaconConfig().ZeroHash[:], newState.FinalizedCheckpoint().Root)
	assert.Equal(t, types.Epoch(0), newState.FinalizedCheckpointEpoch(), "Unexpected finalized epoch")
}
