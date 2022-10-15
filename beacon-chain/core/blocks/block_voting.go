package blocks

import (
	"context"
	"errors"

	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
)

// ProcessBlockVoting is an operation performed on each beacon block
// to collect attestations' consensus.
func ProcessBlockVoting(ctx context.Context, beaconState state.BeaconState, signed block.SignedBeaconBlock) (state.BeaconState, error) {
	if beaconState == nil || beaconState.IsNil() {
		return nil, errors.New("nil state")
	}
	block := signed.Block()
	attestations := block.Body().Attestations()

	//get previous voting result
	preSup, _, err := helpers.GetBlockVotingResults(beaconState.BlockVoting())
	if err != nil {
		return beaconState, err
	}
	// calc total number of voters for block
	var totalAttrs uint64
	committees, err := helpers.CalcSlotCommitteesIndexes(ctx, beaconState, block.Slot())
	if err != nil {
		return beaconState, err
	}
	for _, c := range committees {
		totalAttrs += uint64(len(c))
	}
	candidates := block.Body().Eth1Data().GetCandidates()
	//add item of block voting for the current block
	if err := beaconState.AddBlockVoting(block.ParentRoot(), totalAttrs, candidates); err != nil {
		//if err := beaconState.AddBlockVoting(block.StateRoot(), totalAttrs, candidates); err != nil {
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
	//get post-handling voting result
	postSup, _, err := helpers.GetBlockVotingResults(beaconState.BlockVoting())
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
