package blocks

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/config/params"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
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
	preSup, _, err := GetBlockVotingResults(beaconState.BlockVoting())
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
	postSup, _, err := GetBlockVotingResults(beaconState.BlockVoting())
	if err != nil {
		return beaconState, err
	}

	// remove deprecated voting items
	preSupRoots := BlockVotingsRootsHashArray(preSup)
	postSupRoots := BlockVotingsRootsHashArray(postSup)
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

// GetBlockVotingResults retrieves from BlockVoting supporting results.
func GetBlockVotingResults(blockVoting []*ethpb.BlockVoting) (supported, unsupported []*ethpb.BlockVoting, err error) {
	for _, bv := range blockVoting {
		minSupport, err := blockVotingMinSupport(bv)
		if err != nil {
			return supported, unsupported, err
		}
		if blockVotingCountVotes(bv) >= uint64(minSupport) {
			supported = append(supported, bv)
		} else {
			unsupported = append(unsupported, bv)
		}
	}
	return supported, unsupported, nil
}

// blockVotingMinSupport calc minimal required number of votes for BlockVoting consensus
func blockVotingMinSupport(blockVoting *ethpb.BlockVoting) (int, error) {
	if blockVoting.GetTotalAttesters() == 0 {
		return 0, fmt.Errorf("BlockVoting struct is not fully initialized (root=%#x)", blockVoting.Root)
	}
	minSupport := params.BeaconConfig().BlockVotingMinSupportPrc
	slotValidators := blockVoting.TotalAttesters
	return int(math.Ceil((float64(slotValidators) / 100) * float64(minSupport))), nil
}

// blockVotingCountVotes counts votes of BlockVoting
func blockVotingCountVotes(blockVoting *ethpb.BlockVoting) uint64 {
	count := uint64(0)
	for _, att := range blockVoting.Attestations {
		if helpers.IsAggregated(att) {
			count += att.GetAggregationBits().Count()
		} else {
			count++
		}
	}
	return count
}

// BlockVotingsRootsHashArray returns HashArray of roots of array of BlockVotings.
func BlockVotingsRootsHashArray(blockVoting []*ethpb.BlockVoting) gwatCommon.HashArray {
	roots := make(gwatCommon.HashArray, len(blockVoting))
	for i, bv := range blockVoting {
		roots[i] = gwatCommon.BytesToHash(bv.GetRoot())
	}
	return roots
}
