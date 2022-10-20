package helpers

import (
	"bytes"
	"context"
	"fmt"
	"math"

	types "github.com/prysmaticlabs/eth2-types"
	log "github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/encoding/bytesutil"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
	"sort"
)

type mapVoting map[gwatCommon.Hash]int
type mapPriority map[int]gwatCommon.HashArray
type mapCandidates map[gwatCommon.Hash]gwatCommon.HashArray

// BlockVotingsCalcFinalization calculates finalization sequence by BlockVotings.
func BlockVotingsCalcFinalization(ctx context.Context, state state.BeaconState, blockVotingArr []*ethpb.BlockVoting, lastFinSpine gwatCommon.Hash) (gwatCommon.HashArray, error) {
	var (
		blockVotings    = BlockVotingArrCopy(blockVotingArr)
		supportedVotes  = make([]*ethpb.BlockVoting, 0)
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
				//supportedVotes = append(supportedVotes, ethpb.CopyBlockVoting(bv))
				supportedVotes = append(supportedVotes, BlockVotingCopy(bv))
			}
		}
	}

	log.WithFields(log.Fields{
		"total-BlockVoting":     len(blockVotings),
		"supported-BlockVoting": len(supportedVotes),
	}).Warn("VOTING INFO >>>>")

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

		reductedParts := []gwatCommon.HashArray{candidates}
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
				if candidates.IsEqualTo(rc) {
					intersect := candidates
					key := intersect.Key()
					tabCandidates[key] = intersect
					tabPriority[len(intersect)] = append(tabPriority[len(intersect)], key).Uniq()
					tabVoting[key]++
				}
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

// BlockVotingCopy
func BlockVotingCopy(vote *ethpb.BlockVoting) *ethpb.BlockVoting {
	attestations := make([]*ethpb.Attestation, len(vote.Attestations))
	for i, att := range vote.Attestations {
		attestations[i] = &ethpb.Attestation{
			AggregationBits: att.AggregationBits,
			Data: &ethpb.AttestationData{
				Slot:            att.GetData().GetSlot(),
				CommitteeIndex:  att.GetData().GetCommitteeIndex(),
				BeaconBlockRoot: bytesutil.SafeCopyBytes(att.GetData().BeaconBlockRoot),
				Source: &ethpb.Checkpoint{
					Epoch: att.GetData().GetSource().GetEpoch(),
					Root:  bytesutil.SafeCopyBytes(att.GetData().GetSource().GetRoot()),
				},
				Target: &ethpb.Checkpoint{
					Epoch: att.GetData().GetTarget().GetEpoch(),
					Root:  bytesutil.SafeCopyBytes(att.GetData().GetTarget().GetRoot()),
				},
			},
			Signature: bytesutil.SafeCopyBytes(att.Signature),
		}
	}
	return &ethpb.BlockVoting{
		Root:           bytesutil.SafeCopyBytes(vote.Root),
		TotalAttesters: vote.TotalAttesters,
		Candidates:     bytesutil.SafeCopyBytes(vote.Candidates),
		Attestations:   attestations,
	}
}

// BlockVotingArrCopy
func BlockVotingArrCopy(votes []*ethpb.BlockVoting) []*ethpb.BlockVoting {
	cpy := make([]*ethpb.BlockVoting, len(votes))
	for i, vote := range votes {
		cpy[i] = BlockVotingCopy(vote)
	}
	return cpy
}

// BlockVotingArrSort put BlockVoting array to order to calculate state hash.
func BlockVotingArrStateOrder(votes []*ethpb.BlockVoting) ([]*ethpb.BlockVoting, error) {
	var err error
	cpyAttOrd := make([]*ethpb.BlockVoting, len(votes))
	for i, itm := range votes {
		cpyItm := BlockVotingCopy(itm)
		cpyItm.Attestations, err = AttestationArrSort(cpyItm.Attestations)
		if err != nil {
			return nil, err
		}
		cpyAttOrd[i] = cpyItm
	}
	return BlockVotingArrSort(cpyAttOrd)
}

// BlockVotingArrSort sorts BlockVoting array.
func BlockVotingArrSort(votes []*ethpb.BlockVoting) ([]*ethpb.BlockVoting, error) {
	keys := gwatCommon.HashArray{}
	mapKeyData := map[[32]byte]*ethpb.BlockVoting{}
	for _, itm := range votes {
		k, err := itm.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		mapKeyData[k] = itm
	}
	for k := range mapKeyData {
		keys = append(keys, k)
	}
	keys = keys.Sort()
	sorted := make([]*ethpb.BlockVoting, len(keys))
	for i, k := range keys {
		sorted[i] = mapKeyData[k]
	}
	return sorted, nil
}

// AttestationArrSort sorts attestations array.
func AttestationArrSort(atts []*ethpb.Attestation) ([]*ethpb.Attestation, error) {
	keys := gwatCommon.HashArray{}
	mapKeyData := map[[32]byte]*ethpb.Attestation{}
	for _, itm := range atts {
		k, err := itm.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		mapKeyData[k] = itm
	}
	for k := range mapKeyData {
		keys = append(keys, k)
	}
	keys = keys.Sort()
	sorted := make([]*ethpb.Attestation, len(keys))
	for i, k := range keys {
		sorted[i] = mapKeyData[k]
	}
	return sorted, nil
}

func IndexOfRoot(arrRoots [][]byte, root []byte) int {
	for i, r := range arrRoots {
		if bytes.Equal(r, root) {
			return i
		}
	}
	return -1
}

// PrintBlockVotingArr returns formatted string of BlockVoting array.
func PrintBlockVotingArr(blockVotings []*ethpb.BlockVoting) string {
	str := "["
	for i, bv := range blockVotings {
		str += PrintBlockVoting(bv)
		if i < len(blockVotings)-1 {
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
		if i < len(cands)-1 {
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
		str += fmt.Sprintf("committeeIndex: %d", att.Data.GetCommitteeIndex())
		//str += fmt.Sprintf("beaconBlockRoot: \"%#x\"", att.Data.GetBeaconBlockRoot())
		str += "}"
		if i < len(blockVoting.Attestations)-1 {
			str += ","
		}
	}
	str += fmt.Sprintf("]")
	str += "}"
	return str
}
