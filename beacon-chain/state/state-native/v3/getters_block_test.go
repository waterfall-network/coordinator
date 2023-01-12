package v3

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	testtmpl "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/testing"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

func TestBeaconState_LatestBlockHeader(t *testing.T) {
	testtmpl.VerifyBeaconStateLatestBlockHeader(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconStateBellatrix{})
		},
		func(BH *ethpb.BeaconBlockHeader) (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconStateBellatrix{LatestBlockHeader: BH})
		},
	)
}

func TestBeaconState_BlockRoots(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootsNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconStateBellatrix{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconStateBellatrix{BlockRoots: BR})
		},
	)
}

func TestBeaconState_BlockRootAtIndex(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootAtIndexNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconStateBellatrix{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconStateBellatrix{BlockRoots: BR})
		},
	)
}
