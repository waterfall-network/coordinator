package blockchain

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
)

// NewSlot mimics the implementation of `on_tick` in fork choice consensus spec.
// It resets the proposer boost root in fork choice, and it updates store's justified checkpoint
// if a better checkpoint on the store's finalized checkpoint chain.
// This should only be called at the start of every slot interval.
//
// Spec pseudocode definition:
//
//	# Reset store.proposer_boost_root if this is a new slot
//	if current_slot > previous_slot:
//	    store.proposer_boost_root = Root()
//
//	# Not a new epoch, return
//	if not (current_slot > previous_slot and compute_slots_since_epoch_start(current_slot) == 0):
//	    return
//
//	# Update store.justified_checkpoint if a better checkpoint on the store.finalized_checkpoint chain
//	if store.best_justified_checkpoint.epoch > store.justified_checkpoint.epoch:
//	    finalized_slot = compute_start_slot_at_epoch(store.finalized_checkpoint.epoch)
//	    ancestor_at_finalized_slot = get_ancestor(store, store.best_justified_checkpoint.root, finalized_slot)
//	    if ancestor_at_finalized_slot == store.finalized_checkpoint.root:
//	        store.justified_checkpoint = store.best_justified_checkpoint
func (s *Service) NewSlot(ctx context.Context, slot types.Slot) error {

	go s.updateOptimisticSpineData(ctx, slot)

	// Reset proposer boost root in fork choice.
	if err := s.cfg.ForkChoiceStore.ResetBoostedProposerRoot(ctx); err != nil {
		return errors.Wrap(err, "could not reset boosted proposer root in fork choice")
	}

	// Return if it's not a new epoch.
	if !slots.IsEpochStart(slot) {
		return nil
	}

	// Update store.justified_checkpoint if a better checkpoint on the store.finalized_checkpoint chain
	bj := s.store.BestJustifiedCheckpt()
	if bj == nil {
		return errNilBestJustifiedInStore
	}
	j := s.store.JustifiedCheckpt()
	if j == nil {
		return errNilJustifiedInStore
	}
	f := s.store.FinalizedCheckpt()
	if f == nil {
		return errNilFinalizedInStore
	}
	if bj.Epoch > j.Epoch {
		finalizedSlot, err := slots.EpochStart(f.Epoch)
		if err != nil {
			return err
		}
		r, err := s.ancestor(ctx, bj.Root, finalizedSlot)
		if err != nil {
			return err
		}
		if bytes.Equal(r, f.Root) {
			s.store.SetJustifiedCheckpt(bj)
		}
	}
	return nil

}

func (s *Service) updateOptimisticSpineData(ctx context.Context, slot types.Slot) {
	if s.isSynchronizing() {
		log.WithError(fmt.Errorf("Node syncing to latest head, not ready to respond")).WithFields(logrus.Fields{
			"Syncing": s.isSynchronizing(),
		}).Warn("Optimistic Spines update skipped (synchronizing)")
		return
	}
	if s.IsGwatSynchronizing() {
		log.WithError(fmt.Errorf("GWAT synchronization process is running, not ready to respond")).WithFields(logrus.Fields{
			"Syncing": s.IsGwatSynchronizing(),
		}).Warn("Optimistic Spines update skipped (gwat synchronizing)")
		return
	}

	// calculate the parent block by optimistic spine.
	currHead, err := s.HeadState(ctx)
	if err != nil {
		log.WithError(fmt.Errorf("could not get head state %v", err)).Error("Optimistic Spines update: retrieving of head state failed")
		return
	}

	jCpRoot := bytesutil.ToBytes32(currHead.CurrentJustifiedCheckpoint().Root)
	if currHead.CurrentJustifiedCheckpoint().Epoch == 0 {
		jCpRoot, err = s.cfg.BeaconDB.GenesisBlockRoot(ctx)
		if err != nil {
			log.WithError(errors.Wrap(err, "get genesis root failed")).Error("Optimistic Spines update: retrieving of genesis root failed")
			return
		}
	}

	cpSt, err := s.cfg.StateGen.StateByRoot(ctx, jCpRoot)
	if err != nil {
		log.WithError(fmt.Errorf("could not get head state %v", err)).WithFields(logrus.Fields{
			"jCpRoot": fmt.Sprintf("%#x", jCpRoot),
		}).Error("Optimistic Spines update: retrieving of cp state failed")
		return
	}

	//request optimistic spine
	baseSpine := helpers.GetTerminalFinalizedSpine(cpSt)

	optSpines, err := s.cfg.ExecutionEngineCaller.ExecutionDagGetOptimisticSpines(s.ctx, baseSpine)
	if err != nil {
		errWrap := fmt.Errorf("could not get gwat optSpines: %v", err)
		log.WithError(errWrap).WithFields(logrus.Fields{
			"baseSpine": baseSpine,
		}).Error("Optimistic Spines update: retrieving opt spine failed")
		return
	}
	s.setCacheOptimisticSpines(baseSpine, optSpines)

	log.WithFields(logrus.Fields{
		"slot":      slot,
		"baseSpine": fmt.Sprintf("%#x", baseSpine),
		"opSpines":  optSpines,
	}).Info("Optimistic Spines updated")

	return
}
