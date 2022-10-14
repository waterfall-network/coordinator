package blocks

import (
	"context"
	"errors"
	"math"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/config/params"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
)

// ProcessBlockVoting is an operation performed on each beacon block
// to collect attestations consensus.
func ProcessBlockVoting(ctx context.Context, beaconState state.BeaconState, attestations []*ethpb.Attestation) (state.BeaconState, error) {
	if beaconState == nil || beaconState.IsNil() {
		return nil, errors.New("nil state")
	}
	// todo
	//preFinalization := beaconState.Eth1Data().Finalization
	//preCandidates := beaconState.Eth1Data().Candidates
	//preState := beaconState.Copy()

	for _, att := range attestations {
		if err := beaconState.AppendBlockVotingAtt(att); err != nil {
			return beaconState, err
		}
	}

	supported, unsupported, err := GetBlockVotingRoots(ctx, beaconState, beaconState.BlockVoting())
	if err != nil {
		return beaconState, err
	}

	//todo
	if len(supported) > 0 || len(unsupported) > 0 {

	}

	return beaconState, nil
}

// GetBlockVotingRoots retrieves from BlockVoting roots with enough support.
func GetBlockVotingRoots(ctx context.Context, beaconState state.BeaconState, blockVoting []*ethpb.BlockVoting) (supported, unsupported [][]byte, err error) {
	for _, bv := range blockVoting {
		minSupport, err := blockVotingMinSupport(ctx, beaconState, bv)
		if err != nil {
			return supported, unsupported, err
		}
		if blockVotingCountVotes(bv) >= uint64(minSupport) {
			supported = append(supported, bv.GetRoot())
		} else {
			unsupported = append(unsupported, bv.GetRoot())
		}
	}
	return supported, unsupported, nil
}

// blockVotingMinSupport calc minimal required number of votes for BlockVoting consensus
func blockVotingMinSupport(ctx context.Context, beaconState state.BeaconState, blockVoting *ethpb.BlockVoting) (int, error) {
	minSupport := params.BeaconConfig().BlockVotingMinSupportPrc
	slot, committeeIndex := blockVotingSlotCommitteeIndex(blockVoting)
	committee, err := helpers.BeaconCommitteeFromState(ctx, beaconState, slot, committeeIndex)
	if err != nil {
		return 0, err
	}
	slotValidators := len(committee)
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

// blockVotingSlotCommitteeIndex returns first slot and CommitteeIndex of BlockVoting
func blockVotingSlotCommitteeIndex(blockVoting *ethpb.BlockVoting) (types.Slot, types.CommitteeIndex) {
	minSlot := types.Slot(0)
	minCommitteeIndex := types.CommitteeIndex(0)
	for _, att := range blockVoting.Attestations {
		if minSlot == 0 {
			minSlot = att.Data.Slot
		}
		if att.Data.Slot < minSlot {
			minSlot = att.Data.Slot
		}
		if minCommitteeIndex == 0 {
			minCommitteeIndex = att.Data.CommitteeIndex
		}
		if att.Data.CommitteeIndex < minCommitteeIndex {
			minCommitteeIndex = att.Data.CommitteeIndex
		}
	}
	return minSlot, minCommitteeIndex
}
