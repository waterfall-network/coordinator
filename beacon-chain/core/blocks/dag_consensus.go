package blocks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"sort"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stateutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	attaggregation "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/attestation/aggregation/attestations"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/block"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

type mapVoting map[gwatCommon.Hash]int
type mapPriority map[int]gwatCommon.HashArray
type mapCandidates map[gwatCommon.Hash]gwatCommon.HashArray

// ProcessDagConsensus is an operation performed on each beacon block
// to calculate state transition data related to dag consensus:
// 1. calculate new prefix of spines
// 2. collect attestations and calculate consensus of finalisation.
func ProcessDagConsensus(ctx context.Context, beaconState state.BeaconState, signed block.SignedBeaconBlock) (state.BeaconState, error) {
	if beaconState == nil || beaconState.IsNil() {
		return nil, errors.New("nil state")
	}

	var (
		beaconBlock  = signed.Block()
		attestations = beaconBlock.Body().Attestations()
		candidates   = beaconBlock.Body().Eth1Data().GetCandidates()
		blockVoting  = helpers.BlockVotingArrCopy(beaconState.BlockVoting())
	)

	//calculate prefix
	prefix, unpubChains, err := CalcPrefixAndParentSpines(beaconState.SpineData(), candidates)
	if err != nil {
		return nil, err
	}

	//add item of block voting for the current block
	if len(candidates) > 0 {
		blockVoting = addBlockVoting(blockVoting, beaconBlock.ParentRoot(), beaconBlock.Slot()-1, prefix.ToBytes())
	}
	//append attestations of the current block to block voting
	for _, att := range attestations {
		blockVoting = appendBlockVotingAtt(blockVoting, att)
	}

	//calculation of finalization sequence
	finalization, err := calcFinalization(ctx, beaconState, blockVoting)
	if err != nil {
		return nil, err
	}

	// cutoff the finalization
	prefix = prefix.Difference(finalization)

	// cutoff the finalization & prefix and cast to []*SpinesSeq
	parentSpines := make([]*ethpb.SpinesSeq, 0, len(unpubChains))
	for _, chain := range unpubChains {
		dif := chain.Difference(finalization).Difference(prefix)
		if len(dif) > 0 {
			parentSpines = append(parentSpines, &ethpb.SpinesSeq{Spines: dif.ToBytes()})
		}
	}

	// update Eth1Data of state
	eth1Data := beaconState.Eth1Data()
	eth1Data.Candidates = candidates
	if err = beaconState.SetEth1Data(eth1Data); err != nil {
		return nil, err
	}

	spineData := &ethpb.SpineData{
		Spines:       candidates,
		Prefix:       prefix.ToBytes(),
		Finalization: finalization.ToBytes(),
		CpFinalized:  beaconState.SpineData().CpFinalized,
		ParentSpines: parentSpines,
	}
	if err = beaconState.SetSpineData(spineData); err != nil {
		return nil, err
	}

	// removes BlockVoting with completely finalized candidates
	deprecatedRoots := getBlockVotingsDeprecatedRoots(blockVoting, finalization)
	blockVoting = removeBlockVoting(blockVoting, deprecatedRoots)

	// if it's a new epoch - removes stale BlockVoting.
	if slots.IsEpochStart(beaconBlock.Slot()) {
		cpSlot, err := slots.EpochStart(beaconState.FinalizedCheckpointEpoch())
		if err != nil {
			return nil, err
		}
		staleRoots := getBlockVotingRootsLtSlot(blockVoting, cpSlot)
		blockVoting = removeBlockVoting(blockVoting, staleRoots)

		log.WithFields(logrus.Fields{
			"BlockVoting":      len(blockVoting),
			"StateBlockVoting": len(beaconState.BlockVoting()),
			"staleRoots":       fmt.Sprintf("%#x", staleRoots),
		}).Info("Block Voting processing: removes stale at new epoch")

	}

	blockVoting, err = helpers.BlockVotingArrStateOrder(blockVoting)
	if err != nil {
		return nil, err
	}

	if err := beaconState.SetBlockVoting(blockVoting); err != nil {
		return nil, err
	}
	return beaconState, nil
}

// calcFinalization calculates finalization sequence by BlockVotings.
func calcFinalization(ctx context.Context, beaconState state.BeaconState, blockVotingArr []*ethpb.BlockVoting) (gwatCommon.HashArray, error) {
	var (
		blockVotings    = helpers.BlockVotingArrCopy(blockVotingArr)
		supportedVotes  = make([]*ethpb.BlockVoting, 0)
		candidatesParts = make([][]gwatCommon.HashArray, 0)
		tabPriority     = mapPriority{}
		tabVoting       = mapVoting{}
		tabCandidates   = mapCandidates{}
		slotsToConfirm  = params.BeaconConfig().VotingRequiredSlots
		badVotes        = make([]*ethpb.BlockVoting, 0)
		finalization    = gwatCommon.HashArrayFromBytes(beaconState.SpineData().Finalization)
		lastFinSpine    = helpers.GetTerminalFinalizedSpine(beaconState)
		resFinalization = finalization.Copy()
	)

	for _, bv := range blockVotings {
		// candidates must be uniq
		candidates := gwatCommon.HashArrayFromBytes(bv.Candidates)
		if !candidates.IsUniq() {
			log.WithFields(logrus.Fields{
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
			minSupport, err := BlockVotingMinSupport(ctx, beaconState, slot)
			if err != nil {
				return nil, err
			}
			// if provided enough support for slot adds data as separated item
			if helpers.CountAttestationsVotes(atts) >= uint64(minSupport) {
				supportedVotes = append(supportedVotes, helpers.BlockVotingCopy(bv))
				break
			}
		}
	}

	log.WithFields(logrus.Fields{
		"total-BlockVoting":     len(blockVotings),
		"supported-BlockVoting": len(supportedVotes),
		"VotingRequiredSlots":   params.BeaconConfig().VotingRequiredSlots,
		"SecondsPerSlot":        params.BeaconConfig().SecondsPerSlot,
	}).Info("Voting info")

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
			log.WithFields(logrus.Fields{
				"root":         fmt.Sprintf("%#x", bv.GetRoot()),
				"candidates":   candidates,
				"lastFinSpine": lastFinSpine.Hex(),
			}).Warn("skip: no candidates")
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
	priorities := make([]int, 0)
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

	log.WithFields(logrus.Fields{
		"lastFinSpine": lastFinSpine.Hex(),
		"finalization": tabCandidates[resKey],
	}).Info("Calculation of finalization sequence")

	if resKey != (gwatCommon.Hash{}) {
		resFinalization = append(resFinalization, tabCandidates[resKey]...)
	}

	return resFinalization, nil
}

// BlockVotingMinSupport calc minimal required number of votes for BlockVoting consensus.
func BlockVotingMinSupport(ctx context.Context, state state.BeaconState, slot types.Slot) (int, error) {
	// 50% + 1
	minSupport := params.BeaconConfig().BlockVotingMinSupportPrc
	committees, err := helpers.CalcSlotCommitteesIndexes(ctx, state, slot)
	if err != nil {
		return 0, err
	}
	slotAtts := 0
	for _, cmt := range committees {
		slotAtts += len(cmt)
	}
	val := int(math.Ceil((float64(slotAtts)/100)*float64(minSupport))) + 1
	if val > slotAtts {
		return slotAtts, nil
	}
	return val, nil
}

func CalcPrefixAndParentSpines(stSpineData *ethpb.SpineData, blCandidates []byte) (prefix gwatCommon.HashArray, parentSpines []gwatCommon.HashArray, err error) {
	var (
		parentPrefix                      = gwatCommon.HashArrayFromBytes(stSpineData.Prefix)
		prefixExtension, resPrefix        gwatCommon.HashArray
		parentUnpubChains, resUnpubChains []gwatCommon.HashArray
	)
	spines := gwatCommon.HashArrayFromBytes(blCandidates)
	finalized := gwatCommon.HashArrayFromBytes(stSpineData.Finalization)

	//calc parent unpublished spines
	parentUnpubChains = make([]gwatCommon.HashArray, len(stSpineData.ParentSpines))
	for i, spseq := range stSpineData.ParentSpines {
		chain := gwatCommon.HashArrayFromBytes(spseq.Spines)
		if len(chain) > 0 {
			parentUnpubChains[i] = chain
		}
	}

	// calculate unpublished spines chains
	resUnpubChains = parentUnpubChains
	if len(spines) > 0 {
		// set new spines to the first position
		if len(spines) > 0 {
			resUnpubChains = []gwatCommon.HashArray{spines}
		}
		for _, chain := range parentUnpubChains {
			if len(chain) == 0 {
				continue
			}
			chainDif := chain.Difference(parentPrefix)
			// the first spine of dif-chain must be equal to the first spine
			// otherwise - skip
			if len(chainDif) > 0 && chainDif[0] == spines[0] {
				resUnpubChains = append(resUnpubChains, chain)
			}
		}
	}

	// calculate the new prefix
	prefixExtension, err = helpers.ConsensusCalcPrefix(resUnpubChains)
	if err != nil {
		return nil, nil, err
	}
	resPrefix = append(parentPrefix, prefixExtension...).Uniq()
	resPrefix = resPrefix.Difference(finalized)

	return resPrefix, resUnpubChains, nil
}

// getBlockVotingsDeprecatedRoots returns deprecated roots of BlockVoting.
func getBlockVotingsDeprecatedRoots(blockVoting []*ethpb.BlockVoting, finalization gwatCommon.HashArray) [][]byte {
	mapRoots := map[gwatCommon.Hash][]byte{}
	for _, bv := range blockVoting {
		candidates := gwatCommon.HashArrayFromBytes(bv.GetCandidates())
		if len(candidates) > 0 {
			lastCandidat := candidates[len(candidates)-1]
			if finalization.IndexOf(lastCandidat) > -1 {
				mapRoots[gwatCommon.BytesToHash(bv.GetRoot())] = bv.GetRoot()
			}
		}
	}
	roots := make([][]byte, len(mapRoots))
	i := 0
	for _, rt := range mapRoots {
		roots[i] = rt
		i++
	}
	return roots
}

func getBlockVotingRootsLtSlot(blockVoting []*ethpb.BlockVoting, slot types.Slot) [][]byte {
	mapRoots := map[gwatCommon.Hash][]byte{}
	for _, bv := range blockVoting {
		if bv.Slot <= slot {
			mapRoots[gwatCommon.BytesToHash(bv.GetRoot())] = bv.GetRoot()
		}
	}
	roots := make([][]byte, len(mapRoots))
	i := 0
	for _, rt := range mapRoots {
		roots[i] = rt
		i++
	}
	return roots
}

func isBlockVotingExists(votes []*ethpb.BlockVoting, root []byte) bool {
	for _, itm := range votes {
		if bytes.Equal(itm.Root, root) {
			return true
		}
	}
	return false
}

func addBlockVoting(votes []*ethpb.BlockVoting, root []byte, slot types.Slot, candidates []byte) []*ethpb.BlockVoting {
	cpy := helpers.BlockVotingArrCopy(votes)
	if !isBlockVotingExists(cpy, root) {
		newItem := &ethpb.BlockVoting{
			Root:         root,
			Attestations: []*ethpb.Attestation{},
			Slot:         slot,
			Candidates:   candidates,
		}
		cpy = append(cpy, newItem)
		return cpy
	}
	for _, itm := range cpy {
		if bytes.Equal(itm.Root, root) {
			itm.Slot = slot
			itm.Candidates = candidates
		}
	}
	return votes
}

func appendBlockVotingAtt(votes []*ethpb.BlockVoting, val *ethpb.Attestation) []*ethpb.BlockVoting {
	root := val.GetData().BeaconBlockRoot
	if !isBlockVotingExists(votes, root) {
		return votes
	}
	cpy := helpers.BlockVotingArrCopy(votes)
	for _, itm := range cpy {
		if bytes.Equal(itm.Root, root) {

			log.WithFields(logrus.Fields{
				"atts": helpers.PrintBlockVoting(itm),
				"slot": itm.GetSlot(),
				"root": fmt.Sprintf("%#x", itm.GetRoot()),
			}).Debug("appendBlockVotingAtt op=000")

			valDataRoot, err := val.Data.HashTreeRoot()
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"slot": itm.GetSlot(),
					"root": fmt.Sprintf("%#x", itm.GetRoot()),
				}).Error("append attestation to block voting failed (val HashTreeRoot)")
				return votes
			}
			attsByDataRoot := make(map[[32]byte][]*ethpb.Attestation, len(itm.GetAttestations()))
			for _, att := range itm.GetAttestations() {
				attDataRoot, err := att.Data.HashTreeRoot()
				if err != nil {
					log.WithError(err).WithFields(logrus.Fields{
						"slot": itm.GetSlot(),
						"root": fmt.Sprintf("%#x", itm.GetRoot()),
					}).Error("append attestation to block voting failed (val HashTreeRoot)")
					continue
					//return votes
				}
				attsByDataRoot[attDataRoot] = append(attsByDataRoot[attDataRoot], att)
			}
			if attsByDataRoot[valDataRoot] != nil {
				datts, err := attaggregation.Aggregate(append(attsByDataRoot[valDataRoot], val))
				if err != nil {
					log.WithError(err).WithFields(logrus.Fields{
						"slot": itm.GetSlot(),
						"root": fmt.Sprintf("%#x", itm.GetRoot()),
					}).Error("append attestation to block voting failed (aggregation)")
					return votes
				}
				attsByDataRoot[valDataRoot] = datts
			} else {
				attsByDataRoot[valDataRoot] = []*ethpb.Attestation{val}
			}

			atts := make([]*ethpb.Attestation, 0, len(attsByDataRoot))
			for _, datts := range attsByDataRoot {
				atts = append(atts, datts...)
			}

			ccc := helpers.BlockVotingCopy(itm)
			ccc.Attestations = atts
			log.WithFields(logrus.Fields{
				"BlockVoting": helpers.PrintBlockVoting(ccc),
				"slot":        itm.GetSlot(),
				"root":        fmt.Sprintf("%#x", itm.GetRoot()),
			}).Debug("appendBlockVotingAtt op=1 aggregation")

			atts, err = stateutil.Dedup(atts)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"slot": itm.GetSlot(),
					"root": fmt.Sprintf("%#x", itm.GetRoot()),
				}).Error("append attestation to block voting failed (deduplication)")
				return votes
			}

			ccc.Attestations = atts
			log.WithFields(logrus.Fields{
				"BlockVoting": helpers.PrintBlockVoting(ccc),
				"slot":        itm.GetSlot(),
				"root":        fmt.Sprintf("%#x", itm.GetRoot()),
			}).Debug("appendBlockVotingAtt op=2 deduplication")

			itm.Attestations = atts
		}
	}
	return cpy
}

func removeBlockVoting(votes []*ethpb.BlockVoting, roots [][]byte) []*ethpb.BlockVoting {
	if len(roots) == 0 {
		return votes
	}
	upVotes := make([]*ethpb.BlockVoting, 0)
	for _, itm := range votes {
		if helpers.IndexOfRoot(roots, itm.Root) == -1 {
			upVotes = append(upVotes, itm)
		}
	}
	return upVotes
}
