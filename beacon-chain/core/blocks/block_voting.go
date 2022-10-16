package blocks

import (
	"context"
	"errors"

	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

// ProcessBlockVoting is an operation performed on each beacon block
// to collect attestations' consensus.
func ProcessBlockVoting(ctx context.Context, beaconState state.BeaconState, signed block.SignedBeaconBlock, lastFinSpine gwatCommon.Hash) (state.BeaconState, error) {
	if beaconState == nil || beaconState.IsNil() {
		return nil, errors.New("nil state")
	}
	beaconBlock := signed.Block()
	attestations := beaconBlock.Body().Attestations()

	candidates := beaconBlock.Body().Eth1Data().GetCandidates()
	//add item of block voting for the current block
	if err := beaconState.AddBlockVoting(beaconBlock.ParentRoot(), 0, candidates); err != nil {
		return beaconState, err
	}
	//append attestations of the current block to block voting
	for _, att := range attestations {
		if beaconState.IsBlockVotingExists(att.Data.GetBeaconBlockRoot()) {
			if err := beaconState.AppendBlockVotingAtt(att); err != nil {
				return beaconState, err
			}
		}
	}
	//calculation of finalization sequence
	finalization, err := helpers.BlockVotingsCalcFinalization(ctx, beaconState, beaconState.BlockVoting(), lastFinSpine)
	if err != nil {
		return beaconState, err
	}

	// remove deprecated voting items
	preSupRoots := helpers.BlockVotingsRootsHashArray(preSup)
	postSupRoots := helpers.BlockVotingsRootsHashArray(postSup)
	inRoots := preSupRoots.Intersection(postSupRoots)
	depracatedRoots := make([][]byte, len(inRoots))
	for i, root := range inRoots {
		depracatedRoots[i] = root[:]
	}
	if err := beaconState.RemoveBlockVoting(depracatedRoots); err != nil {
		return beaconState, err
	}

	return beaconState, nil
}
