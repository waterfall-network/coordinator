package v2

import (
	"testing"

	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/testing/require"
)

func TestBeaconState_AppendCurrentEpochAttestations(t *testing.T) {
	s, err := InitializeFromProtoUnsafe(&ethpb.BeaconStateAltair{})
	require.NoError(t, err)
	require.ErrorContains(t, "AppendCurrentEpochAttestations is not supported for hard fork 1 beacon state", s.AppendCurrentEpochAttestations(nil))
}

func TestBeaconState_AppendPreviousEpochAttestations(t *testing.T) {
	s, err := InitializeFromProtoUnsafe(&ethpb.BeaconStateAltair{})
	require.NoError(t, err)
	require.ErrorContains(t, "AppendPreviousEpochAttestations is not supported for hard fork 1 beacon state", s.AppendPreviousEpochAttestations(nil))
}
