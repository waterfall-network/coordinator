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
	if len(blockVoting) > 100 {
		panic("len(blockVoting) > 100")
	}

	//helpers.BlockVotingArrUniq

	//log.WithFields(logrus.Fields{
	//	"ParentRoot":  fmt.Sprintf("%#x", beaconBlock.ParentRoot()),
	//	"BlockVoting": helpers.PrintBlockVotingArr(blockVoting),
	//}).Info("********** ProcessBlockVoting ********** 0000")

	//add item of block voting for the current block
	if len(candidates) > 0 {
		blockVoting = addBlockVoting(blockVoting, beaconBlock.ParentRoot(), uint64(beaconBlock.Slot()-1), candidates)
		//if err := beaconState.AddBlockVoting(beaconBlock.ParentRoot(), 0, candidates); err != nil {
		//	return nil, err
		//}
	}

	//log.WithFields(logrus.Fields{
	//	"ParentRoot":  fmt.Sprintf("%#x", beaconBlock.ParentRoot()),
	//	"BlockVoting": helpers.PrintBlockVotingArr(blockVoting),
	//}).Info("********** ProcessBlockVoting ********** 1111")

	//append attestations of the current block to block voting
	for _, att := range attestations {
		blockVoting = appendBlockVotingAtt(blockVoting, att)
		//if beaconState.IsBlockVotingExists(att.Data.GetBeaconBlockRoot()) {
		//	if err := beaconState.AppendBlockVotingAtt(att); err != nil {
		//		return nil, err
		//	}
		//}
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
	//if err := beaconState.RemoveBlockVoting(deprecatedRoots); err != nil {
	//	return nil, err
	//}

	log.WithFields(logrus.Fields{
		"BlockVoting":     len(blockVoting),
		"deprecatedRoots": fmt.Sprintf("%#x", deprecatedRoots),
	}).Info("********** ProcessBlockVoting ********** 3333")

	// removes stale BlockVoting
	//todo on change epoch only
	cpSlot, err := slots.EpochStart(beaconState.FinalizedCheckpointEpoch())
	if err != nil {
		return nil, err
	}
	staleRoots := getBlockVotingRootsLtSlot(blockVoting, uint64(cpSlot))
	//blockVoting = removeBlockVoting(blockVoting, staleRoots)

	//log.WithFields(logrus.Fields{
	//	"BlockVoting": helpers.PrintBlockVotingArr(blockVoting),
	//	"staleRoots":  fmt.Sprintf("%#x", staleRoots),
	//}).Info("********** ProcessBlockVoting ********** 4444")

	if err := beaconState.SetBlockVoting([]*ethpb.BlockVoting{}); err != nil {
		return nil, err
	}

	for _, bv := range blockVoting {
		if err := beaconState.AppendBlockVoting(bv); err != nil {
			return nil, err
		}
	}

	log.WithFields(logrus.Fields{
		"BlockVoting":      len(blockVoting),
		"StateBlockVoting": len(beaconState.BlockVoting()),
		"deprecatedRoots":  fmt.Sprintf("%#x", deprecatedRoots),
		"staleRoots":       fmt.Sprintf("%#x", staleRoots),
	}).Info("********** ProcessBlockVoting ********** 4444")

	//log.WithError(err).WithFields(logrus.Fields{
	//	"block.slot": signed.Block().Slot(),
	//	//"BlockVoting":     helpers.PrintBlockVotingArr(beaconState.BlockVoting()),
	//	"BlockVoting": len(beaconState.BlockVoting()),
	//	//"cpSlot":          cpSlot,
	//	//"deprecatedRoots": fmt.Sprintf("%#x", deprecatedRoots),
	//	"deprecatedRoots":    len(deprecatedRoots),
	//	"State.Finalization": gwatCommon.HashArrayFromBytes(beaconState.Eth1Data().Finalization),
	//}).Info("--------- ProcessBlockVoting ---------")

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
				//roots = append(roots, bv.GetRoot())
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
		for _, rmRoot := range roots {
			if !bytes.Equal(itm.Root, rmRoot) {
				upVotes = append(upVotes, itm)
			}
		}
	}
	return upVotes
}
