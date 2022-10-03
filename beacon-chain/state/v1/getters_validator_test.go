package v1_test

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	testtmpl "github.com/waterfall-foundation/coordinator/beacon-chain/state/testing"
	v1 "github.com/waterfall-foundation/coordinator/beacon-chain/state/v1"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
)

func TestBeaconState_ValidatorAtIndexReadOnly_HandlesNilSlice(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorAtIndexReadOnlyHandlesNilSlice(t, func() (state.BeaconState, error) {
		return v1.InitializeFromProtoUnsafe(&ethpb.BeaconState{
			Validators: nil,
		})
	})
}
