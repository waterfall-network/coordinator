package blockchain

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/feed"
	statefeed "github.com/waterfall-foundation/coordinator/beacon-chain/core/feed/state"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	doublylinkedtree "github.com/waterfall-foundation/coordinator/beacon-chain/forkchoice/doubly-linked-tree"
	"github.com/waterfall-foundation/coordinator/beacon-chain/forkchoice/protoarray"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/config/features"
	fieldparams "github.com/waterfall-foundation/coordinator/config/fieldparams"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/encoding/bytesutil"
	ethpbv1 "github.com/waterfall-foundation/coordinator/proto/eth/v1"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	"github.com/waterfall-foundation/coordinator/time/slots"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"go.opencensus.io/trace"
)

// UpdateAndSaveHeadWithBalances updates the beacon state head after getting justified balanced from cache.
// This function is only used in spec-tests, it does save the head after updating it.
func (s *Service) UpdateAndSaveHeadWithBalances(ctx context.Context) error {

	log.Info("UpdateAndSaveHeadWithBalances >>>>> 0")

	cp := s.store.JustifiedCheckpt()
	if cp == nil {
		return errors.New("no justified checkpoint")
	}
	balances, err := s.justifiedBalances.get(ctx, bytesutil.ToBytes32(cp.Root))
	if err != nil {
		msg := fmt.Sprintf("could not read balances for state w/ justified checkpoint %#x", cp.Root)
		return errors.Wrap(err, msg)
	}
	headRoot, err := s.updateHead(ctx, balances)
	if err != nil {
		return errors.Wrap(err, "could not update head")
	}
	headBlock, err := s.cfg.BeaconDB.Block(ctx, headRoot)
	if err != nil {
		return err
	}
	headState, err := s.cfg.StateGen.StateByRoot(ctx, headRoot)
	if err != nil {
		return errors.Wrap(err, "could not retrieve head state in DB")
	}
	log.Info("UpdateAndSaveHeadWithBalances >>>>> 11111")

	return s.saveHead(ctx, headRoot, headBlock, headState)
}

// This defines the current chain service's view of head.
type head struct {
	slot  types.Slot              // current head slot.
	root  [32]byte                // current head root.
	block block.SignedBeaconBlock // current head block.
	state state.BeaconState       // current head state.
}

// Determined the head from the fork choice service and saves its new data
// (head root, head block, and head state) to the local service cache.
func (s *Service) updateHead(ctx context.Context, balances []uint64) ([32]byte, error) {
	ctx, span := trace.StartSpan(ctx, "blockChain.updateHead")
	defer span.End()

	// Get head from the fork choice service.
	f := s.store.FinalizedCheckpt()
	if f == nil {
		return [32]byte{}, errNilFinalizedInStore
	}
	j := s.store.JustifiedCheckpt()
	if j == nil {
		return [32]byte{}, errNilJustifiedInStore
	}
	// To get head before the first justified epoch, the fork choice will start with origin root
	// instead of zero hashes.
	headStartRoot := bytesutil.ToBytes32(j.Root)
	if headStartRoot == params.BeaconConfig().ZeroHash {
		headStartRoot = s.originBlockRoot
	}

	// In order to process head, fork choice store requires justified info.
	// If the fork choice store is missing justified block info, a node should
	// re-initiate fork choice store using the latest justified info.
	// This recovers a fatal condition and should not happen in run time.
	if !s.cfg.ForkChoiceStore.HasNode(headStartRoot) {
		jb, err := s.cfg.BeaconDB.Block(ctx, headStartRoot)
		if err != nil {
			return [32]byte{}, err
		}
		if features.Get().EnableForkChoiceDoublyLinkedTree {
			s.cfg.ForkChoiceStore = doublylinkedtree.New(j.Epoch, f.Epoch)
		} else {
			s.cfg.ForkChoiceStore = protoarray.New(j.Epoch, f.Epoch, bytesutil.ToBytes32(f.Root))
		}
		if err := s.insertBlockToForkChoiceStore(ctx, jb.Block(), headStartRoot, f, j); err != nil {
			return [32]byte{}, err
		}
	}

	return s.cfg.ForkChoiceStore.Head(ctx, j.Epoch, headStartRoot, balances, f.Epoch)
}

// This saves head info to the local service cache, it also saves the
// new head root to the DB.
func (s *Service) saveHead(ctx context.Context, headRoot [32]byte, headBlock block.SignedBeaconBlock, headState state.BeaconState) error {
	ctx, span := trace.StartSpan(ctx, "blockChain.saveHead")
	defer span.End()

	// Do nothing if head hasn't changed.
	r, err := s.HeadRoot(ctx)
	if err != nil {
		return err
	}
	if headRoot == bytesutil.ToBytes32(r) {
		return nil
	}
	if err := helpers.BeaconBlockIsNil(headBlock); err != nil {
		return err
	}
	if headState == nil || headState.IsNil() {
		return errors.New("cannot save nil head state")
	}

	// If the head state is not available, just return nil.
	// There's nothing to cache
	if !s.cfg.BeaconDB.HasStateSummary(ctx, headRoot) {
		return nil
	}

	// A chain re-org occurred, so we fire an event notifying the rest of the services.
	headSlot := s.HeadSlot()
	newHeadSlot := headBlock.Block().Slot()
	oldHeadRoot := s.headRoot()
	oldStateRoot := s.headBlock().Block().StateRoot()
	newStateRoot := headBlock.Block().StateRoot()
	if bytesutil.ToBytes32(headBlock.Block().ParentRoot()) != bytesutil.ToBytes32(r) {
		log.WithFields(logrus.Fields{
			"newSlot": fmt.Sprintf("%d", newHeadSlot),
			"oldSlot": fmt.Sprintf("%d", headSlot),
		}).Info("Chain reorg occurred")
		absoluteSlotDifference := slots.AbsoluteValueSlotDifference(newHeadSlot, headSlot)
		isOptimistic, err := s.IsOptimistic(ctx)
		if err != nil {
			return errors.Wrap(err, "could not check if node is optimistically synced")
		}
		s.cfg.StateNotifier.StateFeed().Send(&feed.Event{
			Type: statefeed.Reorg,
			Data: &ethpbv1.EventChainReorg{
				Slot:                newHeadSlot,
				Depth:               absoluteSlotDifference,
				OldHeadBlock:        oldHeadRoot[:],
				NewHeadBlock:        headRoot[:],
				OldHeadState:        oldStateRoot,
				NewHeadState:        newStateRoot,
				Epoch:               slots.ToEpoch(newHeadSlot),
				ExecutionOptimistic: isOptimistic,
			},
		})

		if err := s.saveOrphanedAtts(ctx, bytesutil.ToBytes32(r)); err != nil {
			return err
		}

		reorgCount.Inc()
	}

	var finalizedSeq gwatCommon.HashArray

	if !s.isSync() {
		cp := headState.CurrentJustifiedCheckpoint()
		cpState, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(cp.Root))
		if err != nil {
			log.WithError(errors.Wrapf(err, "could not get checkpoint state for epoch=%d root=%x", cp.Epoch, cp.GetRoot())).Error("save head")
			return errors.Wrapf(err, "could not get checkpoint state for epoch=%d root=%x", cp.Epoch, cp.GetRoot())
		}
		if cpState == nil || cpState.IsNil() {
			log.WithError(errors.Wrapf(err, "checkpoint's state not found for epoch=%d root=%x", cp.Epoch, cp.GetRoot())).Error("save head")
			return errors.Wrapf(err, "checkpoint's state not found for epoch=%d root=%x", cp.Epoch, cp.GetRoot())
		}

		cpFin := gwatCommon.HashArrayFromBytes(cpState.Eth1Data().Finalization)
		headFin := gwatCommon.HashArrayFromBytes(headState.Eth1Data().Finalization)
		skip := headFin.IsEqualTo(cpFin)
		if !skip {
			baseSpine, finalizing, err := s.collectFinalizationParams(ctx, headBlock, headState)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"finalizing": finalizing,
					"baseSpine":  baseSpine.Hex(),
				}).Warn("saveHead: get finalization params failed")
				return errors.Wrap(err, "saveHead: get finalization params failed")
			}

			lfSpine, err := s.cfg.ExecutionEngineCaller.ExecutionDagFinalize(ctx, finalizing, &baseSpine)
			fSeq := append(gwatCommon.HashArray{baseSpine}, finalizing...)
			if err != nil || lfSpine == nil {

				log.WithError(err).WithFields(logrus.Fields{
					"finalizing": finalizing,
					"baseSpine":  baseSpine.Hex(),
					"lfSpine":    lfSpine,
				}).Warn("saveHead: finalization failed")
				return errors.Wrap(err, "saveHead: gwat finalization failed")
			}
			for _, h := range fSeq {
				finalizedSeq = append(finalizedSeq, h)
				if h == *lfSpine {
					break
				}
			}

			if len(finalizedSeq) == 0 {
				err = errors.New("lf spine is invalid")
				log.WithError(err).WithFields(logrus.Fields{
					"finalizationSeq": fSeq,
					"lfSpine":         *lfSpine,
					"isValid":         fSeq.Has(*lfSpine),
				}).Warn("saveHead: finalization failed")
				return errors.Wrap(err, "saveHead: gwat finalization failed")
			}

			log.WithFields(logrus.Fields{
				"finalized":        finalizing,
				"baseSpine":        baseSpine.Hex(),
				"lfSpine":          lfSpine.Hex(),
				"finalizationSeq":  fSeq,
				"finalizedSeq":     finalizedSeq,
				"isFullyFinalized": fSeq.IsEqualTo(finalizedSeq),
			}).Info("save head: finalization success")
		}
	}

	// Cache the new head info.
	s.setHead(headRoot, headBlock, headState)

	// Save the new head root to DB.
	if err := s.cfg.BeaconDB.SaveHeadBlockRoot(ctx, headRoot); err != nil {
		return errors.Wrap(err, "could not save head root in DB")
	}

	//update checkpoint of FinalizedSpines cache
	cpFin := headState.FinalizedCheckpoint()
	cpState, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(cpFin.Root))
	if err != nil {
		log.WithError(errors.Wrapf(err, "Cache finalized spines: could not get checkpoint state for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot())).Error("save head")
		return errors.Wrapf(err, "Cache finalized spines: could not get checkpoint state for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot())
	}
	if cpState == nil || cpState.IsNil() {
		log.WithError(errors.Wrapf(err, "Cache finalized spines: checkpoint's state not found for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot())).Error("save head")
		return errors.Wrapf(err, "Cache finalized spines: checkpoint's state not found for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot())
	}
	cpFinSeq := gwatCommon.HashArrayFromBytes(cpState.Eth1Data().Finalization)
	s.SetFinalizedSpinesCheckpoint(cpFinSeq[len(cpFinSeq)-1])

	//update FinalizedSpines cache
	s.AddFinalizedSpines(finalizedSeq)

	// Forward an event capturing a new chain head over a common event feed
	// done in a goroutine to avoid blocking the critical runtime main routine.
	go func() {
		if err := s.notifyNewHeadEvent(ctx, newHeadSlot, headState, newStateRoot, headRoot[:]); err != nil {
			log.WithError(err).Error("Could not notify event feed of new chain head")
		}
	}()

	return nil
}

// This gets called to update canonical root mapping. It does not save head block
// root in DB. With the inception of initial-sync-cache-state flag, it uses finalized
// check point as anchors to resume sync therefore head is no longer needed to be saved on per slot basis.
func (s *Service) saveHeadNoDB(ctx context.Context, b block.SignedBeaconBlock, r [32]byte, hs state.BeaconState) error {
	if err := helpers.BeaconBlockIsNil(b); err != nil {
		return err
	}
	cachedHeadRoot, err := s.HeadRoot(ctx)
	if err != nil {
		return errors.Wrap(err, "could not get head root from cache")
	}
	if bytes.Equal(r[:], cachedHeadRoot) {
		return nil
	}

	s.setHeadInitialSync(r, b.Copy(), hs)
	return nil
}

// This sets head view object which is used to track the head slot, root, block and state.
func (s *Service) setHead(root [32]byte, block block.SignedBeaconBlock, state state.BeaconState) {
	s.headLock.Lock()
	defer s.headLock.Unlock()

	stRoot, err := state.HashTreeRoot(s.ctx)
	if err != nil {

	}
	log.WithError(err).WithFields(logrus.Fields{
		"block.Slot":   block.Block().Slot(),
		"block.Parent": fmt.Sprintf("%#x", block.Block().ParentRoot()),
		"state.Slot":   state.Slot(),
		"state.Root":   fmt.Sprintf("%#x", stRoot),
	}).Info("setHead >>>>> 11111")

	// This does a full copy of the block and state.
	s.head = &head{
		slot:  block.Block().Slot(),
		root:  root,
		block: block.Copy(),
		state: state.Copy(),
	}
}

// This sets head view object which is used to track the head slot, root, block and state. The method
// assumes that state being passed into the method will not be modified by any other alternate
// caller which holds the state's reference.
func (s *Service) setHeadInitialSync(root [32]byte, block block.SignedBeaconBlock, state state.BeaconState) {
	s.headLock.Lock()
	defer s.headLock.Unlock()

	// This does a full copy of the block only.
	s.head = &head{
		slot:  block.Block().Slot(),
		root:  root,
		block: block.Copy(),
		state: state,
	}
}

// This returns the head slot.
// This is a lock free version.
func (s *Service) headSlot() types.Slot {
	return s.head.slot
}

// This returns the head root.
// It does a full copy on head root for immutability.
// This is a lock free version.
func (s *Service) headRoot() [32]byte {
	if s.head == nil {
		return params.BeaconConfig().ZeroHash
	}

	return s.head.root
}

// This returns the head block.
// It does a full copy on head block for immutability.
// This is a lock free version.
func (s *Service) headBlock() block.SignedBeaconBlock {
	return s.head.block.Copy()
}

// This returns the head state.
// It does a full copy on head state for immutability.
// This is a lock free version.
func (s *Service) headState(ctx context.Context) state.BeaconState {
	_, span := trace.StartSpan(ctx, "blockChain.headState")
	defer span.End()

	return s.head.state.Copy()
}

// This returns the genesis validators root of the head state.
// This is a lock free version.
func (s *Service) headGenesisValidatorsRoot() [32]byte {
	return bytesutil.ToBytes32(s.head.state.GenesisValidatorsRoot())
}

// This returns the validator referenced by the provided index in
// the head state.
// This is a lock free version.
func (s *Service) headValidatorAtIndex(index types.ValidatorIndex) (state.ReadOnlyValidator, error) {
	return s.head.state.ValidatorAtIndexReadOnly(index)
}

// This returns the validator index referenced by the provided pubkey in
// the head state.
// This is a lock free version.
func (s *Service) headValidatorIndexAtPubkey(pubKey [fieldparams.BLSPubkeyLength]byte) (types.ValidatorIndex, bool) {
	return s.head.state.ValidatorIndexByPubkey(pubKey)
}

// Returns true if head state exists.
// This is the lock free version.
func (s *Service) hasHeadState() bool {
	return s.head != nil && s.head.state != nil
}

// Notifies a common event feed of a new chain head event. Called right after a new
// chain head is determined, set, and saved to disk.
func (s *Service) notifyNewHeadEvent(
	ctx context.Context,
	newHeadSlot types.Slot,
	newHeadState state.BeaconState,
	newHeadStateRoot,
	newHeadRoot []byte,
) error {
	previousDutyDependentRoot := s.originBlockRoot[:]
	currentDutyDependentRoot := s.originBlockRoot[:]

	var previousDutyEpoch types.Epoch
	currentDutyEpoch := slots.ToEpoch(newHeadSlot)
	if currentDutyEpoch > 0 {
		previousDutyEpoch = currentDutyEpoch.Sub(1)
	}
	currentDutySlot, err := slots.EpochStart(currentDutyEpoch)
	if err != nil {
		return errors.Wrap(err, "could not get duty slot")
	}
	previousDutySlot, err := slots.EpochStart(previousDutyEpoch)
	if err != nil {
		return errors.Wrap(err, "could not get duty slot")
	}
	if currentDutySlot > 0 {
		currentDutyDependentRoot, err = helpers.BlockRootAtSlot(newHeadState, currentDutySlot-1)
		if err != nil {
			return errors.Wrap(err, "could not get duty dependent root")
		}
	}
	if previousDutySlot > 0 {
		previousDutyDependentRoot, err = helpers.BlockRootAtSlot(newHeadState, previousDutySlot-1)
		if err != nil {
			return errors.Wrap(err, "could not get duty dependent root")
		}
	}
	isOptimistic, err := s.IsOptimistic(ctx)
	if err != nil {
		return errors.Wrap(err, "could not check if node is optimistically synced")
	}
	s.cfg.StateNotifier.StateFeed().Send(&feed.Event{
		Type: statefeed.NewHead,
		Data: &ethpbv1.EventHead{
			Slot:                      newHeadSlot,
			Block:                     newHeadRoot,
			State:                     newHeadStateRoot,
			EpochTransition:           slots.IsEpochStart(newHeadSlot),
			PreviousDutyDependentRoot: previousDutyDependentRoot,
			CurrentDutyDependentRoot:  currentDutyDependentRoot,
			ExecutionOptimistic:       isOptimistic,
		},
	})
	return nil
}

// This saves the attestations inside the beacon block with respect to root `orphanedRoot` back into the
// attestation pool. It also filters out the attestations that is one epoch older as a
// defense so invalid attestations don't flow into the attestation pool.
func (s *Service) saveOrphanedAtts(ctx context.Context, orphanedRoot [32]byte) error {
	if !features.Get().CorrectlyInsertOrphanedAtts {
		return nil
	}

	orphanedBlk, err := s.cfg.BeaconDB.Block(ctx, orphanedRoot)
	if err != nil {
		return err
	}

	if orphanedBlk == nil || orphanedBlk.IsNil() {
		return errors.New("orphaned block can't be nil")
	}

	for _, a := range orphanedBlk.Block().Body().Attestations() {
		// Is the attestation one epoch older.
		if a.Data.Slot+params.BeaconConfig().SlotsPerEpoch < s.CurrentSlot() {
			continue
		}
		if helpers.IsAggregated(a) {
			if err := s.cfg.AttPool.SaveAggregatedAttestation(a); err != nil {
				return err
			}
		} else {
			if err := s.cfg.AttPool.SaveUnaggregatedAttestation(a); err != nil {
				return err
			}
		}
		saveOrphanedAttCount.Inc()
	}

	return nil
}

// This saves head info to the local service cache, it also saves the
// new head root to the DB.
func (s *Service) collectFinalizationParams(
	ctx context.Context,
	headBlock block.SignedBeaconBlock,
	headState state.BeaconState,
) (baseSpine gwatCommon.Hash, finalizationSeq gwatCommon.HashArray, err error) {
	if headState == nil || headState.IsNil() {
		return baseSpine, finalizationSeq, errors.New("Collect finalization params: nil head state received")
	}
	//update checkpoint of FinalizedSpines cache
	checkpoint := headState.FinalizedCheckpoint()
	cpSlot, err := slots.EpochStart(checkpoint.Epoch)
	if err != nil {
		return baseSpine, finalizationSeq, err
	}
	finalizedSpines := s.GetFinalizedSpines()
	var currRoot [32]byte
	currState := headState
	currBlock := headBlock
	for {
		currFinalization := gwatCommon.HashArrayFromBytes(currState.Eth1Data().Finalization)
		intersect := finalizedSpines.SequenceIntersection(currFinalization)
		if len(intersect) == 0 {
			//finalizationSeq = append(finalizationSeq, currFinalization...)
			finalizationSeq = append(currFinalization, finalizationSeq...)
		} else {
			baseSpine = intersect[len(intersect)-1]
			add := false
			for _, h := range currFinalization {
				if add {
					//finalizationSeq = append(finalizationSeq, h)
					finalizationSeq = append(gwatCommon.HashArray{h}, finalizationSeq...)
				}
				if h == baseSpine {
					add = true
				}
			}
			//update FinalizedSpines cache
			s.SetFinalizedSpinesHead(baseSpine)
			break
		}
		//set next block root as current
		currRoot = bytesutil.ToBytes32(currBlock.Block().ParentRoot())
		if currRoot == params.BeaconConfig().ZeroHash {
			return baseSpine, finalizationSeq, nil
		}
		//set next block as current
		currBlock, err = s.cfg.BeaconDB.Block(s.ctx, currRoot)
		if err != nil {
			return baseSpine, finalizationSeq, err
		}
		//set next state as current
		currState, err = s.cfg.StateGen.StateByRoot(ctx, currRoot)
		if err != nil {
			err = errors.Wrapf(err, "could not get parent state for root=%x", currRoot)
			log.WithError(err).Error("Collect finalization params")
			return baseSpine, finalizationSeq, err
		}
		if currState == nil || currState.IsNil() {
			err = errors.Errorf("retrieved nil parent state for root=%x", currRoot)
			log.WithError(err).Error("Collect finalization params")
			return baseSpine, finalizationSeq, err
		}
		// if reach finalized checkpoint slot
		if currBlock.Block().Slot() < cpSlot {
			err = errors.New("Collect finalization params: failed")
			log.WithError(err).Error("Collect finalization params")
			return baseSpine, finalizationSeq, err
		}
	}
	return baseSpine, finalizationSeq, nil
}
