package blocks

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
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
	//add item of block voting for the current block
	if err := beaconState.AddBlockVoting(beaconBlock.ParentRoot(), 0, candidates); err != nil {
		return nil, err
	}
	//append attestations of the current block to block voting
	for _, att := range attestations {
		if beaconState.IsBlockVotingExists(att.Data.GetBeaconBlockRoot()) {
			if err := beaconState.AppendBlockVotingAtt(att); err != nil {
				return nil, err
			}
		}
	}
	//calculation of finalization sequence
	finalization, err := helpers.BlockVotingsCalcFinalization(ctx, beaconState, beaconState.BlockVoting(), lastFinSpine)
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

	// removes BlockVoting with completely finalized candidates
	deprecatedRoots := getBlockVotingsDeprecatedRoots(beaconState.BlockVoting(), finalization)
	if err := beaconState.RemoveBlockVoting(deprecatedRoots); err != nil {
		return nil, err
	}

	////TODO on epoch changed: removes stale BlockVoting

	log.WithError(err).WithFields(logrus.Fields{
		"block.slot": signed.Block().Slot(),
		//"BlockVoting":     helpers.PrintBlockVotingArr(beaconState.BlockVoting()),
		"BlockVoting": len(beaconState.BlockVoting()),
		//"cpSlot":          cpSlot,
		"deprecatedRoots": fmt.Sprintf("%#x", deprecatedRoots),
		//"staleRoots-len":  len(staleRoots),
	}).Info("--------- ProcessBlockVoting ---------")

	return beaconState, nil
}

// getBlockVotingsDeprecatedRoots returns deprecated roots of BlockVoting.
func getBlockVotingsDeprecatedRoots(blockVoting []*ethpb.BlockVoting, finalization gwatCommon.HashArray) [][]byte {
	roots := [][]byte{}
	for _, bv := range blockVoting {
		candidates := gwatCommon.HashArrayFromBytes(bv.GetCandidates())
		if len(candidates) > 0 {
			lastCandidat := candidates[len(candidates)-1]
			if finalization.IndexOf(lastCandidat) > -1 {
				roots = append(roots, bv.GetRoot())
			}
		}
	}
	return roots
}
