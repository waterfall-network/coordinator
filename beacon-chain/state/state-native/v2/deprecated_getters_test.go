package v2

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func TestBeaconState_CurrentEpochAttestations(t *testing.T) {
	s := &BeaconState{}
	_, err := s.CurrentEpochAttestations()
	require.ErrorContains(t, "CurrentEpochAttestations is not supported for hard fork 1 beacon state", err)
}

func TestBeaconState_PreviousEpochAttestations(t *testing.T) {
	s := &BeaconState{}
	_, err := s.PreviousEpochAttestations()
	require.ErrorContains(t, "PreviousEpochAttestations is not supported for hard fork 1 beacon state", err)
}
