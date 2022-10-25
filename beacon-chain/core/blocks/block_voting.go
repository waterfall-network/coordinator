package blocks

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state/stateutil"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	"github.com/waterfall-foundation/coordinator/time/slots"
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
	blockVoting := helpers.BlockVotingArrCopy(beaconState.BlockVoting())

	//todo rm
	if len(blockVoting) > 160 {
		log.WithFields(logrus.Fields{
			"blockSlot":   signed.Block().Slot(),
			"BlockVoting": helpers.PrintBlockVotingArr(blockVoting),
		}).Error("********** ProcessBlockVoting ********** len(blockVoting) > 160")
	}

	//add item of block voting for the current block
	if len(candidates) > 0 {
		blockVoting = addBlockVoting(blockVoting, beaconBlock.ParentRoot(), uint64(beaconBlock.Slot()-1), candidates)
	}

	//append attestations of the current block to block voting
	for _, att := range attestations {
		blockVoting = appendBlockVotingAtt(blockVoting, att)
	}

	log.WithFields(logrus.Fields{
		"len(blockVoting)": len(blockVoting),
		"BlockVoting":      helpers.PrintBlockVotingArr(blockVoting),
	}).Info("********** ProcessBlockVoting ********** 2222")

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
	eth1Data.Finalization = finalization.ToBytes()
	if err := beaconState.SetEth1Data(eth1Data); err != nil {
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"len(blockVoting)": len(blockVoting),
	}).Info("********** ProcessBlockVoting ********** 0022-999999")

	// removes BlockVoting with completely finalized candidates
	deprecatedRoots := getBlockVotingsDeprecatedRoots(blockVoting, finalization)
	blockVoting = removeBlockVoting(blockVoting, deprecatedRoots)

	log.WithFields(logrus.Fields{
		"BlockVoting":      len(blockVoting),
		"StateBlockVoting": len(beaconState.BlockVoting()),
		"deprecatedRoots":  fmt.Sprintf("%#x", deprecatedRoots),
	}).Info("********** ProcessBlockVoting ********** 4444")

	// if it's a new epoch - removes stale BlockVoting.
	if slots.IsEpochStart(beaconBlock.Slot()) {
		cpSlot, err := slots.EpochStart(beaconState.FinalizedCheckpointEpoch())
		if err != nil {
			return nil, err
		}
		staleRoots := getBlockVotingRootsLtSlot(blockVoting, uint64(cpSlot))
		blockVoting = removeBlockVoting(blockVoting, staleRoots)

		log.WithFields(logrus.Fields{
			"BlockVoting":      len(blockVoting),
			"StateBlockVoting": len(beaconState.BlockVoting()),
			"staleRoots":       fmt.Sprintf("%#x", staleRoots),
		}).Info("********** ProcessBlockVoting ********** 5555 new epoch:removes stale")

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

func getBlockVotingRootsLtSlot(blockVoting []*ethpb.BlockVoting, slot uint64) [][]byte {
	mapRoots := map[gwatCommon.Hash][]byte{}
	for _, bv := range blockVoting {
		if bv.TotalAttesters <= slot {
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

func addBlockVoting(votes []*ethpb.BlockVoting, root []byte, slot uint64, candidates []byte) []*ethpb.BlockVoting {
	cpy := helpers.BlockVotingArrCopy(votes)
	if !isBlockVotingExists(cpy, root) {
		newItem := &ethpb.BlockVoting{
			Root:           root,
			Attestations:   []*ethpb.Attestation{},
			TotalAttesters: slot,
			Candidates:     candidates,
		}
		cpy = append(cpy, newItem)
		return cpy
	}
	for _, itm := range cpy {
		if bytes.Equal(itm.Root, root) {
			itm.TotalAttesters = slot
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
			atts, err := stateutil.Dedup(append(itm.GetAttestations(), val))
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"slot": itm.GetTotalAttesters(),
					"root": fmt.Sprintf("%#x", itm.GetRoot()),
				}).Error("append attestation to block voting failed")
				return votes
			}
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
