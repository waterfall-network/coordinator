package blocks

import (
	"context"
	"errors"

	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
)

// ProcessBlockVoting is an operation performed on each beacon block
// to collect attestations consensus.
func ProcessBlockVoting(_ context.Context, beaconState state.BeaconState, attestations []*ethpb.Attestation) (state.BeaconState, error) {
	if beaconState == nil || beaconState.IsNil() {
		return nil, errors.New("nil state")
	}

	for _, att := range attestations {
		if err := beaconState.AppendBlockVoting(att); err != nil {
			return nil, err
		}
	}
	return beaconState, nil
}
