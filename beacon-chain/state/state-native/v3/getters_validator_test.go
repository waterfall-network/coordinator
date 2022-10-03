package v3_test

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	v3 "github.com/waterfall-foundation/coordinator/beacon-chain/state/state-native/v3"
	testtmpl "github.com/waterfall-foundation/coordinator/beacon-chain/state/testing"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
)

func TestBeaconState_ValidatorAtIndexReadOnly_HandlesNilSlice(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorAtIndexReadOnlyHandlesNilSlice(t, func() (state.BeaconState, error) {
		return v3.InitializeFromProtoUnsafe(&ethpb.BeaconStateBellatrix{
			Validators: nil,
		})
	})
}
