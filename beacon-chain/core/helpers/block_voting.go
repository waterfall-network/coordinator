package helpers

import (
	"fmt"
	"math"

	"github.com/waterfall-foundation/coordinator/config/params"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

// GetBlockVotingResults retrieves from BlockVoting supporting results.
func GetBlockVotingResults(blockVoting []*ethpb.BlockVoting) (supported, unsupported []*ethpb.BlockVoting, err error) {
	for _, bv := range blockVoting {
		minSupport, err := BlockVotingMinSupport(bv)
		if err != nil {
			return supported, unsupported, err
		}
		if BlockVotingCountVotes(bv) >= uint64(minSupport) {
			supported = append(supported, bv)
		} else {
			unsupported = append(unsupported, bv)
		}
	}
	return supported, unsupported, nil
}

// BlockVotingMinSupport calc minimal required number of votes for BlockVoting consensus
func BlockVotingMinSupport(blockVoting *ethpb.BlockVoting) (int, error) {
	if blockVoting.GetTotalAttesters() == 0 {
		return 0, fmt.Errorf("BlockVoting struct is not properly initialized properly (root=%#x)", blockVoting.Root)
	}
	minSupport := params.BeaconConfig().BlockVotingMinSupportPrc
	slotValidators := blockVoting.TotalAttesters
	return int(math.Ceil((float64(slotValidators) / 100) * float64(minSupport))), nil
}

// BlockVotingCountVotes counts votes of BlockVoting
func BlockVotingCountVotes(blockVoting *ethpb.BlockVoting) uint64 {
	count := uint64(0)
	for _, att := range blockVoting.Attestations {
		if IsAggregated(att) {
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

// PrintBlockVotingArr returns formatted string of BlockVoting array.
func PrintBlockVotingArr(blockVotings []*ethpb.BlockVoting) string {
	str := "["
	for i, bv := range blockVotings {
		str += PrintBlockVoting(bv)
		if i < len(blockVotings)-2 {
			str += ","
		}
	}
	str += "]"
	return str
}

// PrintBlockVoting returns formatted string of BlockVoting.
func PrintBlockVoting(blockVoting *ethpb.BlockVoting) string {
	candStr := "["
	cands := gwatCommon.HashArrayFromBytes(blockVoting.GetCandidates())
	for i, c := range cands {
		candStr += fmt.Sprintf("\"%#x\"", c)
		if i < len(cands)-2 {
			candStr += ","
		}
	}
	candStr += "]"

	str := "{"
	str += fmt.Sprintf("root: \"%#x\",", blockVoting.Root)
	str += fmt.Sprintf("candidates: %s,", candStr)
	str += fmt.Sprintf("totalAttesters: %d,", blockVoting.GetTotalAttesters())
	str += fmt.Sprintf("attestations: [")
	for i, att := range blockVoting.Attestations {
		str += "{"
		str += fmt.Sprintf("aggrBitLen: %d,", att.GetAggregationBits().Count())
		str += fmt.Sprintf("aggrBitCount: %d,", att.GetAggregationBits().Count())
		str += fmt.Sprintf("slot: %d,", att.Data.GetSlot())
		str += fmt.Sprintf("committeeIndex: %d,", att.Data.GetCommitteeIndex())
		str += fmt.Sprintf("beaconBlockRoot: \"%#x\",", att.Data.GetBeaconBlockRoot())
		str += "}"
		if i < len(blockVoting.Attestations)-2 {
			str += ","
		}
	}
	str += fmt.Sprintf("]")
	str += "}"
	return str
}
