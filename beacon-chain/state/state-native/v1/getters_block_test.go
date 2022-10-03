package v1

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	testtmpl "github.com/waterfall-foundation/coordinator/beacon-chain/state/testing"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
)

func TestBeaconState_LatestBlockHeader(t *testing.T) {
	testtmpl.VerifyBeaconStateLatestBlockHeader(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconState{})
		},
		func(BH *ethpb.BeaconBlockHeader) (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconState{LatestBlockHeader: BH})
		},
	)
}

func TestBeaconState_BlockRoots(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootsNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconState{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconState{BlockRoots: BR})
		},
	)
}

func TestBeaconState_BlockRootAtIndex(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootAtIndexNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconState{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconState{BlockRoots: BR})
		},
	)
}
