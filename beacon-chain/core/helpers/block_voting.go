package helpers

import (
	"context"
	"fmt"
	types "github.com/prysmaticlabs/eth2-types"
	log "github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/config/params"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
	"math"
	"sort"
)

type mapVoting map[gwatCommon.Hash]int
type mapPriority map[int]gwatCommon.HashArray
type mapCandidates map[gwatCommon.Hash]gwatCommon.HashArray

// BlockVotingsCalcFinalization calculates finalization sequence by BlockVotings.
func BlockVotingsCalcFinalization(ctx context.Context, state state.BeaconState, blockVotings []*ethpb.BlockVoting, lastFinSpine gwatCommon.Hash) (gwatCommon.HashArray, error) {
	var (
		supportedVotes  = []*ethpb.BlockVoting{}
		candidatesParts = [][]gwatCommon.HashArray{}
		tabPriority     = mapPriority{}
		tabVoting       = mapVoting{}
		tabCandidates   = mapCandidates{}
		slotsToConfirm  = 3
		badVotes        = []*ethpb.BlockVoting{}
	)
	for _, bv := range blockVotings {
		// candidates must be uniq
		candidates := gwatCommon.HashArrayFromBytes(bv.Candidates)
		if !candidates.IsUniq() {
			log.WithFields(log.Fields{
				"root":         fmt.Sprintf("%#x", bv.GetRoot()),
				"candidates":   candidates,
				"lastFinSpine": lastFinSpine.Hex(),
			}).Warn("skip: bad candidates (not uniq)")
			badVotes = append(badVotes, bv)
			continue
		}

		// handle data for each attestations' slot
		mapSlotAtt := map[types.Slot][]*ethpb.Attestation{}
		for _, att := range bv.GetAttestations() {
			mapSlotAtt[att.GetData().GetSlot()] = append(mapSlotAtt[att.GetData().GetSlot()], att)
		}
		for slot, atts := range mapSlotAtt {
			minSupport, err := BlockVotingMinSupport(ctx, state, slot)
			if err != nil {
				return nil, err
			}
			// if provided enough support for slot adds data as separated item
			if CountAttestationsVotes(atts) >= uint64(minSupport) {
				supportedVotes = append(supportedVotes, ethpb.CopyBlockVoting(bv))
			}
		}
	}

	// handle supported votes
	for _, bv := range supportedVotes {
		candidates := gwatCommon.HashArrayFromBytes(bv.Candidates)
		//exclude finalized spines
		fullLen := len(candidates)
		if i := candidates.IndexOf(lastFinSpine); i >= 0 {
			candidates = candidates[i+1:]
		}
		// if all current candidates handled
		if len(candidates) == 0 && fullLen > len(candidates) {
			log.WithFields(log.Fields{
				"root":         fmt.Sprintf("%#x", bv.GetRoot()),
				"candidates":   candidates,
				"lastFinSpine": lastFinSpine.Hex(),
			}).Warn("skip: all candidates is finalized")
			badVotes = append(badVotes, bv)
			continue
		}

		reductedParts := []gwatCommon.HashArray{}
		// reduction of sequence up to single item
		for i := len(candidates) - 1; i > 0; i-- {
			reduction := candidates[:i]
			reductedParts = append(reductedParts, reduction)
		}
		candidatesParts = append(candidatesParts, reductedParts)
	}

	//calculate voting params
	for i, candidatesList := range candidatesParts {
		var restList []gwatCommon.HashArray
		restParts := candidatesParts[i+1:]
		for _, rp := range restParts {
			restList = append(restList, rp...)
		}
		for _, candidates := range candidatesList {
			for _, rc := range restList {
				intersect := candidates.SequenceIntersection(rc)
				key := intersect.Key()
				tabCandidates[key] = intersect
				tabPriority[len(intersect)] = append(tabPriority[len(intersect)], key).Uniq()
				tabVoting[key]++
			}
		}
	}

	//sort by priority
	priorities := []int{}
	for p := range tabPriority {
		priorities = append(priorities, p)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(priorities)))

	// find voting result
	resKey := gwatCommon.Hash{}
	for _, p := range priorities {
		// select by max number of vote which satisfies the condition
		// of min required number of votes
		maxVotes := 0
		for _, key := range tabPriority[p] {
			votes := tabVoting[key]
			if votes >= slotsToConfirm && votes > maxVotes {
				resKey = key
			}
		}
		if resKey != (gwatCommon.Hash{}) {
			break
		}
	}

	log.WithFields(log.Fields{
		"lastFinSpine": lastFinSpine.Hex(),
		"finalization": tabCandidates[resKey],
	}).Info("Calculation of finalization sequence")

	if resKey == (gwatCommon.Hash{}) {
		return gwatCommon.HashArray{}, nil
	}
	return tabCandidates[resKey], nil

}

// BlockVotingMinSupport calc minimal required number of votes for BlockVoting consensus
func BlockVotingMinSupport(ctx context.Context, state state.BeaconState, slot types.Slot) (int, error) {
	minSupport := params.BeaconConfig().BlockVotingMinSupportPrc
	committees, err := CalcSlotCommitteesIndexes(ctx, state, slot)
	if err != nil {
		return 0, err
	}
	slotAtts := 0
	for _, cmt := range committees {
		slotAtts += len(cmt)
	}
	return int(math.Ceil((float64(slotAtts) / 100) * float64(minSupport))), nil
}

// BlockVotingCountVotes counts votes of BlockVoting
func CountAttestationsVotes(attestations []*ethpb.Attestation) uint64 {
	count := uint64(0)
	for _, att := range attestations {
		if IsAggregated(att) {
			count += att.GetAggregationBits().Count()
		} else {
			count++
		}
	}
	return count
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
