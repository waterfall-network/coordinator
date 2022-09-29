package initialsync

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/beacon-chain/db/filters"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/encoding/bytesutil"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	"github.com/waterfall-foundation/coordinator/time/slots"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
	gwatTypes "github.com/waterfall-foundation/gwat/core/types"
)

func (s *Service) IsInitSync(ctx context.Context) bool {
	return s.isInitSynchronizing
}

// headSync implements head-sync procedure with gwat node.
//
// Step 1 - Head sync readiness to provide checking of readiness of both sides to head-sync main procedure,
//
// Step 2 - Head Sync (main procedure) to sync both sides from finalized epoch to head
// and make sure of the consistence of nodes.
func (s *Service) HeadSync(ctx context.Context, force bool) error {
	if !force && s.IsInitSync(ctx) {
		return nil
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Step 1
	if err := s.execHeadSyncReady(ctx); err != nil {
		return err
	}
	// Step 2
	return s.execHeadSync(ctx)
}

// execHeadSyncReady sync from head to best known finalized epoch.
func (s *Service) execHeadSyncReady(ctx context.Context) error {
	s.isInitSynchronizing = true
	ticker := time.NewTicker(time.Duration(params.BeaconConfig().HeadSyncReadyIntervalMs) * time.Millisecond)
	defer func() {
		ticker.Stop()
	}()

	log.WithField("HeadSyncReadyIntervalMs", fmt.Sprintf("%d", params.BeaconConfig().HeadSyncReadyIntervalMs)).Info("Head sync readiness starts ...")

	for {
		checkpoint := s.cfg.Chain.CurrentJustifiedCheckpt()
		if checkpoint == nil {
			log.Error("Head sync readiness: no checkpoint")
			continue
		}
		cpRoot := bytesutil.ToBytes32(checkpoint.Root)
		//if cpRoot == params.BeaconConfig().ZeroHash {
		//	log.WithField("checkpoint.Root", fmt.Sprintf("%x", checkpoint.Root)).Error("Head sync readiness: checkpoint.Root empty")
		//	continue
		//}

		cpState, err := s.cfg.StateGen.StateByRoot(ctx, cpRoot)
		if err != nil {
			log.WithField("cpState", cpState).WithError(err).Error("Head sync readiness: error")
			continue
		}

		log.WithFields(logrus.Fields{
			"checkpoint.Root":  fmt.Sprintf("%x", checkpoint.Root),
			"checkpoint.Epoch": checkpoint.Epoch,
			"cpState.Slot":     cpState.Slot(),
		}).Info("Head sync readiness: checkpoint")

		creators, err := s.GetCreators(ctx, cpState, cpState.Slot())
		if err != nil {
			log.WithField("cpState", cpState).WithError(err).Error("Head sync readiness: error (get creators)")
			continue
		}

		syncParam := &gwatTypes.ConsensusInfo{
			Slot:       uint64(cpState.Slot()),
			Creators:   creators,
			Finalizing: gwatCommon.HashArrayFromBytes(cpState.Eth1Data().Finalization),
		}

		logFields := logrus.Fields{
			"checkpoint.Epoch": checkpoint.Epoch,
			"checkpoint.Root":  fmt.Sprintf("%x", checkpoint.Root),
			"Slot":             syncParam.Slot,
			"Creators":         syncParam.Creators,
			"Finalizing":       syncParam.Finalizing,
		}

		log.WithFields(logFields).Info("Head sync readiness: params")

		isReady, err := s.cfg.ExecutionEngineCaller.ExecutionDagHeadSyncReady(ctx, syncParam)

		log.WithField("isReady", isReady).WithError(err).Info("Head sync readiness: result")
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("Head sync readiness: error")
		}
		if isReady {
			s.headSyncCp = checkpoint
			log.WithFields(logFields).Info("Head sync readiness: success")
			return nil
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			log.Debug("Head sync readiness: context closed, exiting routine")
			return nil
		}
	}
}

// execHeadSync Sync (main procedure) to sync both sides from finalized epoch to head
// and make sure of the consistence of nodes.
func (s *Service) execHeadSync(ctx context.Context) error {
	defer func() {
		s.isInitSynchronizing = false
	}()
	syncParams := []gwatTypes.ConsensusInfo{}

	cpState, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(s.headSyncCp.Root))
	if err != nil {
		log.WithField("cpState", cpState).WithError(err).Error("Head sync readiness: error")
		return err
	}
	//cpState := s.headSyncState

	log.WithFields(logrus.Fields{
		"checkpoint.Root":  fmt.Sprintf("%x", s.headSyncCp.Root),
		"checkpoint.Epoch": s.headSyncCp.Epoch,
		"cpState.Slot":     cpState.Slot(),
	}).Info("Head sync main: checkpoint")

	startSlot := cpState.Slot()
	endSlot := s.cfg.Chain.HeadSlot()
	//endSlot := s.cfg.Chain.CurrentSlot()

	filter := filters.NewFilter().SetStartSlot(startSlot).SetEndSlot(endSlot) //.SetSlotStep(0)
	blks, roots, err := s.cfg.DB.Blocks(ctx, filter)

	log.WithFields(logrus.Fields{
		"blks":      len(blks),
		"startSlot": startSlot,
		"endSlot":   endSlot,
		"HeadSlot":  s.cfg.Chain.HeadSlot(),
	}).Info("Head sync main: get blocks")

	if err != nil {
		log.WithError(err).Error("Head sync main: error (get blocks)")
		return err
	}
	// handle genesis case
	if startSlot == 0 {
		genBlock, genRoot, err := s.retrieveGenesisBlock(ctx)
		if err != nil {
			log.WithError(err).Error("Head sync main: error (get genesis block)")
			return err
		}
		blks = append([]block.SignedBeaconBlock{genBlock}, blks...)
		roots = append([][32]byte{genRoot}, roots...)
	}
	// Filter and sort our retrieved blocks, so that
	// we only return valid sets of blocks.
	blks, roots, err = s.dedupBlocksAndRoots(blks, roots)
	if err != nil {
		log.WithError(err).Error("Head sync main: error (get valid sets of blocks)")
		return err
	}
	blks, _ = s.sortBlocksAndRoots(blks, roots)
	for _, b := range blks {

		log.WithFields(logrus.Fields{
			"slot": b.Block().Slot(),
		}).Info("Head sync main: block iter")

		if b == nil || b.IsNil() || b.Block().IsNil() {
			continue
		}

		slot := b.Block().Slot()
		bState := cpState
		if bState == nil {
			log.WithField("state", bState).Error("Head sync main: state not found")

			bState, err = s.cfg.DB.State(ctx, bytesutil.ToBytes32(b.Block().StateRoot()))
			log.WithField("state", bState).Error("Head sync main: state not found 2222222")
			if err != nil {
				log.WithField("state", bState).WithError(err).Error("Head sync main: error 222222222")
			}
			panic("state not found")
		}

		creators, err := s.GetCreators(ctx, bState, slot)
		if err != nil {
			log.WithField("slot", slot).WithError(err).Error("Head sync main: error (get head creators)")
			return err
		}

		syncParams = append(syncParams,
			gwatTypes.ConsensusInfo{
				Slot:       uint64(slot),
				Creators:   creators,
				Finalizing: gwatCommon.HashArrayFromBytes(b.Block().Body().Eth1Data().Finalization),
			},
		)
	}

	logFields := logrus.Fields{
		"startSlot": startSlot,
		"endSlot":   endSlot,
		"len(data)": len(syncParams),
	}

	log.WithFields(logFields).Info("Head sync main: syncParams")

	isReady, err := s.cfg.ExecutionEngineCaller.ExecutionDagHeadSync(ctx, syncParams)

	log.WithField("isReady", isReady).WithError(err).Info("Head sync main: result")
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Head sync main: error")
	}
	if isReady {
		log.WithFields(logFields).Info("Head sync main: success")
		return nil
	}

	return nil
}

// This defines the current chain service's view of creators.
type creatorsAssignment struct {
	assignment map[types.Slot][]gwatCommon.Address // creators' assignment by slot.
	lock       sync.RWMutex
}

// GetCreators returns creators assignments for current slot.
func (s *Service) GetCreators(ctx context.Context, state state.BeaconState, slot types.Slot) ([]gwatCommon.Address, error) {
	s.creators.lock.RLock()
	defer s.creators.lock.RUnlock()

	// retrieve creators assignments from cache
	if s.creators.assignment != nil && s.creators.assignment[slot] != nil {
		return s.creators.assignment[slot], nil
	}

	// calculate creators assignments
	epoch := slots.ToEpoch(slot)
	creatorsAssig, err := helpers.CalcCreatorsAssignments(ctx, state, epoch)
	if err != nil {
		return []gwatCommon.Address{}, err
	}
	//cache result
	s.creators.assignment = creatorsAssig
	return s.creators.assignment[slot], nil
}

func (s *Service) retrieveGenesisBlock(ctx context.Context) (block.SignedBeaconBlock, [32]byte, error) {
	genBlock, err := s.cfg.DB.GenesisBlock(ctx)
	if err != nil {
		return nil, [32]byte{}, err
	}
	genRoot, err := genBlock.Block().HashTreeRoot()
	if err != nil {
		return nil, [32]byte{}, err
	}
	return genBlock, genRoot, nil
}

// A type to represent beacon blocks and roots which have methods
// which satisfy the Interface in `Sort` so that this type can
// be sorted in ascending order.
type sortedObj struct {
	blks  []block.SignedBeaconBlock
	roots [][32]byte
}

// Less reports whether the element with index i must sort before the element with index j.
func (s sortedObj) Less(i, j int) bool {
	return s.blks[i].Block().Slot() < s.blks[j].Block().Slot()
}

// Swap swaps the elements with indexes i and j.
func (s sortedObj) Swap(i, j int) {
	s.blks[i], s.blks[j] = s.blks[j], s.blks[i]
	s.roots[i], s.roots[j] = s.roots[j], s.roots[i]
}

// Len is the number of elements in the collection.
func (s sortedObj) Len() int {
	return len(s.blks)
}

// removes duplicates from provided blocks and roots.
func (_ *Service) dedupBlocksAndRoots(blks []block.SignedBeaconBlock, roots [][32]byte) ([]block.SignedBeaconBlock, [][32]byte, error) {
	if len(blks) != len(roots) {
		return nil, nil, errors.New("input blks and roots are diff lengths")
	}

	// Remove duplicate blocks received
	rootMap := make(map[[32]byte]bool, len(blks))
	newBlks := make([]block.SignedBeaconBlock, 0, len(blks))
	newRoots := make([][32]byte, 0, len(roots))
	for i, r := range roots {
		if rootMap[r] {
			continue
		}
		rootMap[r] = true
		newRoots = append(newRoots, roots[i])
		newBlks = append(newBlks, blks[i])
	}
	return newBlks, newRoots, nil
}

func (_ *Service) dedupRoots(roots [][32]byte) [][32]byte {
	newRoots := make([][32]byte, 0, len(roots))
	rootMap := make(map[[32]byte]bool, len(roots))
	for i, r := range roots {
		if rootMap[r] {
			continue
		}
		rootMap[r] = true
		newRoots = append(newRoots, roots[i])
	}
	return newRoots
}

// sort the provided blocks and roots in ascending order. This method assumes that the size of
// block slice and root slice is equal.
func (_ *Service) sortBlocksAndRoots(blks []block.SignedBeaconBlock, roots [][32]byte) ([]block.SignedBeaconBlock, [][32]byte) {
	obj := sortedObj{
		blks:  blks,
		roots: roots,
	}
	sort.Sort(obj)
	return obj.blks, obj.roots
}
