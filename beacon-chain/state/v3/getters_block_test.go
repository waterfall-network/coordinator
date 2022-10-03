package v3

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
			return InitializeFromProto(&ethpb.BeaconStateBellatrix{})
		},
		func(BH *ethpb.BeaconBlockHeader) (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconStateBellatrix{LatestBlockHeader: BH})
		},
	)
}

func TestBeaconState_BlockRoots(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRoots(
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
	testtmpl.VerifyBeaconStateBlockRootAtIndex(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconStateBellatrix{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProto(&ethpb.BeaconStateBellatrix{BlockRoots: BR})
		},
	)
}
