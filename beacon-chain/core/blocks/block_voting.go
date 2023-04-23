package blocks

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stateutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	attaggregation "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/attestation/aggregation/attestations"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/block"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
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
	blockVoting := helpers.BlockVotingArrCopy(beaconState.BlockVoting())

	//add item of block voting for the current block
	if len(candidates) > 0 {
		blockVoting = addBlockVoting(blockVoting, beaconBlock.ParentRoot(), beaconBlock.Slot()-1, candidates)
	}

	//append attestations of the current block to block voting
	for _, att := range attestations {
		blockVoting = appendBlockVotingAtt(blockVoting, att)
	}

	log.WithFields(logrus.Fields{
		"len(blockVoting)": len(blockVoting),
		"BlockVoting":      helpers.PrintBlockVotingArr(blockVoting),
	}).Info("Block Voting processing")

	//calculation of finalization sequence
	finalization, err := helpers.BlockVotingsCalcFinalization(ctx, beaconState, blockVoting, lastFinSpine)
	if err != nil {
		return nil, err
	}

	//if finalization sequence empty - set last finalized spine
	if len(finalization) == 0 {
		finalization = gwatCommon.HashArray{lastFinSpine}
	}
	// update Eth1Data of state
	eth1Data := beaconState.Eth1Data()
	eth1Data.Candidates = candidates
	if err := beaconState.SetEth1Data(eth1Data); err != nil {
		return nil, err
	}

	spineData := beaconState.SpineData()
	spineData.Finalization = finalization.ToBytes()
	if err := beaconState.SetSpineData(spineData); err != nil {
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
			}).Info("??? appendBlockVotingAtt ??? 000")

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
			}).Info("??? appendBlockVotingAtt ??? 111 aggregation")

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
			}).Info("??? appendBlockVotingAtt ??? 222 deduplication")

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
