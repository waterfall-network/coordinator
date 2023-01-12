package v2_test

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	testtmpl "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/testing"
	v2 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v2"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

func TestBeaconState_ValidatorAtIndexReadOnly_HandlesNilSlice(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorAtIndexReadOnlyHandlesNilSlice(t, func() (state.BeaconState, error) {
		return v2.InitializeFromProtoUnsafe(&ethpb.BeaconStateAltair{
			Validators: nil,
		})
	})
}
