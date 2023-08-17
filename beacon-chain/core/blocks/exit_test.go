package blocks_test

import (
	"context"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
)

func TestProcessVoluntaryExits_NotActiveLongEnoughToExit(t *testing.T) {
	exits := []*ethpb.VoluntaryExit{{ValidatorIndex: 0, Epoch: 0}}
	registry := []*ethpb.Validator{
		{
			ExitEpoch:      params.BeaconConfig().FarFutureEpoch,
			ActivationHash: make([]byte, 32),
			ExitHash:       make([]byte, 32),
			WithdrawalOps:  make([]*ethpb.WithdrawalOp, 0),
		},
	}
	state, err := v1.InitializeFromProto(&ethpb.BeaconState{
		Validators: registry,
		Slot:       10,
	})
	require.NoError(t, err)
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			VoluntaryExits: exits,
		},
	}

	want := "validator has not been active long enough to exit"
	_, err = blocks.ProcessVoluntaryExits(context.Background(), state, b.Block.Body.VoluntaryExits)
	assert.ErrorContains(t, want, err)
}

func TestProcessVoluntaryExits_ExitAlreadySubmitted(t *testing.T) {
	exits := []*ethpb.VoluntaryExit{{Epoch: 10}}
	registry := []*ethpb.Validator{
		{
			ExitEpoch:      10,
			ActivationHash: (params.BeaconConfig().ZeroHash)[:],
			ExitHash:       (params.BeaconConfig().ZeroHash)[:],
			WithdrawalOps:  []*ethpb.WithdrawalOp{},
		},
	}
	state, err := v1.InitializeFromProto(&ethpb.BeaconState{
		Validators: registry,
		Slot:       0,
	})
	require.NoError(t, err)
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			VoluntaryExits: exits,
		},
	}

	want := "validator with index 0 has already submitted an exit, which will take place at epoch: 10"
	_, err = blocks.ProcessVoluntaryExits(context.Background(), state, b.Block.Body.VoluntaryExits)
	assert.ErrorContains(t, want, err)
}

func TestProcessVoluntaryExits_AppliesCorrectStatus(t *testing.T) {
	exits := []*ethpb.VoluntaryExit{{
		ValidatorIndex: 0,
		Epoch:          0,
	},
	}
	registry := []*ethpb.Validator{
		{
			ExitEpoch:       params.BeaconConfig().FarFutureEpoch,
			ActivationEpoch: 0,
			ActivationHash:  (params.BeaconConfig().ZeroHash)[:],
			ExitHash:        (params.BeaconConfig().ZeroHash)[:],
			WithdrawalOps:   []*ethpb.WithdrawalOp{},
		},
	}
	state, err := v1.InitializeFromProto(&ethpb.BeaconState{
		Validators: registry,
		Fork: &ethpb.Fork{
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
		},
		Slot: params.BeaconConfig().SlotsPerEpoch * 5,
	})
	require.NoError(t, err)
	err = state.SetSlot(state.Slot() + params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().ShardCommitteePeriod)))
	require.NoError(t, err)

	priv, err := bls.RandKey()
	require.NoError(t, err)

	val, err := state.ValidatorAtIndex(0)
	require.NoError(t, err)
	val.PublicKey = priv.PublicKey().Marshal()
	require.NoError(t, state.UpdateValidatorAtIndex(0, val))

	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			VoluntaryExits: exits,
		},
	}

	newState, err := blocks.ProcessVoluntaryExits(context.Background(), state, b.Block.Body.VoluntaryExits)
	require.NoError(t, err, "Could not process exits")
	newRegistry := newState.Validators()
	if newRegistry[0].ExitEpoch != helpers.ActivationExitEpoch(types.Epoch(state.Slot()/params.BeaconConfig().SlotsPerEpoch)) {
		t.Errorf("Expected validator exit epoch to be %d, got %d",
			helpers.ActivationExitEpoch(types.Epoch(state.Slot()/params.BeaconConfig().SlotsPerEpoch)), newRegistry[0].ExitEpoch)
	}
}
