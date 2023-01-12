package v1

import (
	"context"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stateutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// computeFieldRoots returns the hash tree root computations of every field in
// the beacon state as a list of 32 byte roots.
func computeFieldRoots(ctx context.Context, state *ethpb.BeaconState) ([][]byte, error) {
	return stateutil.ComputeFieldRootsWithHasherPhase0(ctx, state)
}
