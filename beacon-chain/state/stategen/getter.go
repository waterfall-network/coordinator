package stategen

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"go.opencensus.io/trace"
)

var ErrNoDataForSlot = errors.New("cannot retrieve data for slot")

// HasState returns true if the state exists in cache or in DB.
func (s *State) HasState(ctx context.Context, blockRoot [32]byte) (bool, error) {
	has, err := s.HasStateInCache(ctx, blockRoot)
	if err != nil {
		return false, err
	}
	if has {
		return true, nil
	}
	return s.beaconDB.HasState(ctx, blockRoot), nil
}

// HasStateInCache returns true if the state exists in cache.
func (s *State) HasStateInCache(ctx context.Context, blockRoot [32]byte) (bool, error) {
	if s.hotStateCache.has(blockRoot) {
		return true, nil
	}
	_, has, err := s.epochBoundaryStateCache.getByRoot(blockRoot)
	if err != nil {
		return false, err
	}
	return has, nil
}

// StateByRootIfCachedNoCopy retrieves a state using the input block root only if the state is already in the cache
func (s *State) StateByRootIfCachedNoCopy(blockRoot [32]byte) state.BeaconState {
	if !s.hotStateCache.has(blockRoot) {
		return nil
	}
	bState := s.hotStateCache.getWithoutCopy(blockRoot)
	return bState
}

// StateByRoot retrieves the state using input block root.
func (s *State) StateByRoot(ctx context.Context, blockRoot [32]byte) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "stateGen.StateByRoot")
	defer span.End()

	// Genesis case. If block root is zero hash, short circuit to use genesis cachedState stored in DB.
	if blockRoot == params.BeaconConfig().ZeroHash {
		return s.beaconDB.GenesisState(ctx)
	}
	return s.loadStateByRoot(ctx, blockRoot)
}

// StateByRootInitialSync retrieves the state from the DB for the initial syncing phase.
// It assumes initial syncing using a block list rather than a block tree hence the returned
// state is not copied.
// It invalidates cache for parent root because pre state will get mutated.
// Do not use this method for anything other than initial syncing purpose or block tree is applied.
func (s *State) StateByRootInitialSync(ctx context.Context, blockRoot [32]byte) (state.BeaconState, error) {
	// Genesis case. If block root is zero hash, short circuit to use genesis state stored in DB.
	if blockRoot == params.BeaconConfig().ZeroHash {
		return s.beaconDB.GenesisState(ctx)
	}

	// To invalidate cache for parent root because pre state will get mutated.
	defer s.hotStateCache.delete(blockRoot)

	if s.hotStateCache.has(blockRoot) {
		return s.hotStateCache.getWithoutCopy(blockRoot), nil
	}

	cachedInfo, ok, err := s.epochBoundaryStateCache.getByRoot(blockRoot)
	if err != nil {
		return nil, err
	}
	if ok {
		return cachedInfo.state, nil
	}

	startState, err := s.LastAncestorState(ctx, blockRoot)
	if err != nil {
		return nil, errors.Wrap(err, "could not get ancestor state")
	}
	if startState == nil || startState.IsNil() {
		return nil, errUnknownState
	}
	summary, err := s.stateSummary(ctx, blockRoot)
	if err != nil {
		return nil, errors.Wrap(err, "could not get state summary")
	}
	if startState.Slot() == summary.Slot {
		return startState, nil
	}

	blks, err := s.LoadBlocks(ctx, startState.Slot()+1, summary.Slot, bytesutil.ToBytes32(summary.Root))
	if err != nil {
		return nil, errors.Wrap(err, "could not load blocks")
	}
	startState, err = s.ReplayBlocks(ctx, startState, blks, summary.Slot)
	if err != nil {
		return nil, errors.Wrap(err, "could not replay blocks")
	}

	return startState, nil
}

// This returns the state summary object of a given block root, it first checks the cache
// then checks the DB. An error is returned if state summary object is nil.
func (s *State) stateSummary(ctx context.Context, blockRoot [32]byte) (*ethpb.StateSummary, error) {
	var summary *ethpb.StateSummary
	var err error

	summary, err = s.beaconDB.StateSummary(ctx, blockRoot)
	if err != nil {
		return nil, err
	}

	if summary == nil {
		return s.RecoverStateSummary(ctx, blockRoot)
	}
	return summary, nil
}

// RecoverStateSummary recovers state summary object of a given block root by using the saved block in DB.
func (s *State) RecoverStateSummary(ctx context.Context, blockRoot [32]byte) (*ethpb.StateSummary, error) {
	if s.beaconDB.HasBlock(ctx, blockRoot) {
		b, err := s.beaconDB.Block(ctx, blockRoot)
		if err != nil {
			return nil, err
		}
		summary := &ethpb.StateSummary{Slot: b.Block().Slot(), Root: blockRoot[:]}
		if err := s.beaconDB.SaveStateSummary(ctx, summary); err != nil {
			return nil, err
		}
		return summary, nil
	}
	return nil, errors.New("could not find block in DB")
}

// DeleteStateFromCaches deletes the state from the caches.
func (s *State) DeleteStateFromCaches(_ context.Context, blockRoot [32]byte) error {
	s.hotStateCache.delete(blockRoot)
	return s.epochBoundaryStateCache.delete(blockRoot)
}

// This loads a beacon state from either the cache or DB then replay blocks up the requested block root.
func (s *State) loadStateByRoot(ctx context.Context, blockRoot [32]byte) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "stateGen.loadStateByRoot")
	defer span.End()

	defer func(start time.Time) {
		duration := time.Since(start).Microseconds()
		log.WithFields(logrus.Fields{
			"μs":   duration,
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: duration 0")
	}(time.Now())

	// First, it checks if the state exists in hot state cache.
	cachedState := s.hotStateCache.get(blockRoot)
	if cachedState != nil && !cachedState.IsNil() {
		log.WithFields(logrus.Fields{
			"0slot": cachedState.Slot(),
			"root":  fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: from hot cache")
		return cachedState, nil
	}

	// check if the state exists in sync state cache.
	cachedState = s.syncStateCache.get(blockRoot)
	if cachedState != nil && !cachedState.IsNil() {
		log.WithFields(logrus.Fields{
			"0slot": cachedState.Slot(),
			"root":  fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: from sync cache")
		return cachedState, nil
	}

	// Second, it checks if the state exits in epoch boundary state cache.
	cachedInfo, ok, err := s.epochBoundaryStateCache.getByRoot(blockRoot)
	if err != nil {
		return nil, err
	}
	if ok {
		log.WithFields(logrus.Fields{
			"0slot": cachedInfo.state.Slot(),
			"root":  fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: from epoch boundary cache")
		return cachedInfo.state, nil
	}

	// check the state exists in replay cache
	resState, err := s.loadStateCache.Get(ctx, blockRoot)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Error("Load state by root: replay cache get failed")
		return nil, err
	}
	if resState != nil {
		log.WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: from replay cache get")
		return resState, nil
	}

	// Short cut if the cachedState is already in the DB.
	if s.beaconDB.HasState(ctx, blockRoot) {
		log.WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: from DB")
		return s.beaconDB.State(ctx, blockRoot)
	}

	// if same request is already in progress - waite result
	resState, err = s.loadStateCache.GetWhenReady(ctx, blockRoot)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Error("Load state by root: replay cache wait failed")
		return nil, err
	}
	if resState != nil {
		log.WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: from replay wait cache")
		return resState, nil
	}
	if err = s.loadStateCache.MarkInProgress(blockRoot); err != nil {
		if errors.Is(err, cache.ErrAlreadyInProgress) {
			resState, err = s.loadStateCache.GetWhenReady(ctx, blockRoot)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"root": fmt.Sprintf("%#x", blockRoot),
				}).Error("Load state by root: replay cache already progress")
				return nil, err
			}
			if resState == nil || resState.IsNil() {
				err = errors.New("replay cache was in progress and resolved nil")
				log.WithError(err).WithFields(logrus.Fields{
					"root": fmt.Sprintf("%#x", blockRoot),
				}).Error("Load state by root: replay cache resolved nil")
				return nil, err
			}
			log.WithFields(logrus.Fields{
				"root": fmt.Sprintf("%#x", blockRoot),
			}).Info("Load state by root: from replay cache 2")
			return resState, nil
		}
		log.WithError(err).WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Error("Load state by root: replay cache start failed")
		return nil, err
	}
	defer func() {
		if err := s.loadStateCache.MarkNotInProgress(blockRoot); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"root": fmt.Sprintf("%#x", blockRoot),
			}).Error("Load state by root: replay cache failed to mark in progress")
		}
	}()

	log.WithFields(logrus.Fields{
		"root": fmt.Sprintf("%#x", blockRoot),
	}).Info("Load state by root: replay blocks start")

	summary, err := s.stateSummary(ctx, blockRoot)
	if err != nil {
		return nil, errors.Wrap(err, "could not get state summary")
	}
	targetSlot := summary.Slot

	// Since the requested state is not in caches, start replaying using the last available ancestor state which is
	// retrieved using input block's parent root.
	startState, err := s.LastAncestorState(ctx, blockRoot)
	if err != nil {
		return nil, errors.Wrap(err, "could not get ancestor state")
	}
	if startState == nil || startState.IsNil() {
		return nil, errUnknownBoundaryState
	}

	// Return state early if we are retrieving it from our finalized state cache.
	if startState.Slot() == targetSlot {
		return startState, nil
	}

	blks, err := s.LoadBlocks(ctx, startState.Slot()+1, targetSlot, bytesutil.ToBytes32(summary.Root))
	if err != nil {
		return nil, errors.Wrap(err, "could not load blocks for hot state using root")
	}

	replayBlockCount.Observe(float64(len(blks)))

	resState, err = s.ReplayBlocks(ctx, startState, blks, targetSlot)
	if err != nil {
		return resState, err
	}
	// cache result of the ReplayBlocks method
	if err = s.loadStateCache.Put(ctx, blockRoot, resState); err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Error("Load state by root: replay cache put failed")
	}
	return resState, nil
}

// SyncStateByRoot retrieves the state using input block root.
// checks state from hotStateCache is not mutated.
func (s *State) SyncStateByRoot(ctx context.Context, blockRoot [32]byte) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "stateGen.SyncStateByRoot")
	defer span.End()

	// Genesis case. If block root is zero hash, short circuit to use genesis cachedState stored in DB.
	if blockRoot == params.BeaconConfig().ZeroHash {
		return s.beaconDB.GenesisState(ctx)
	}
	return s.loadSyncStateByRoot(ctx, blockRoot)
}

// This loads a beacon state from either the cache or DB then replay blocks up the requested block root.
// checks state from hotStateCache is not mutated.
func (s *State) loadSyncStateByRoot(ctx context.Context, blockRoot [32]byte) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "stateGen.loadSyncStateByRoot")
	defer span.End()

	defer func(start time.Time) {
		duration := time.Since(start).Microseconds()
		log.WithFields(logrus.Fields{
			"μs":   duration,
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: duration 1")
	}(time.Now())

	//// check if the state exists in sync state cache.
	//cachedState := s.hotStateCache.get(blockRoot)
	//if cachedState != nil && !cachedState.IsNil() {
	//	log.WithFields(logrus.Fields{
	//		"0slot": cachedState.Slot(),
	//		"root": fmt.Sprintf("%#x", blockRoot),
	//	}).Info("Load state by root: from hot cache")
	//	return cachedState, nil
	//}

	// check if the state exists in sync state cache.
	cachedState := s.syncStateCache.get(blockRoot)
	if cachedState != nil && !cachedState.IsNil() {
		log.WithFields(logrus.Fields{
			"0slot": cachedState.Slot(),
			"root":  fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: from sync cache")
		return cachedState, nil
	}

	// Second, it checks if the state exits in epoch boundary state cache.
	cachedInfo, ok, err := s.epochBoundaryStateCache.getByRoot(blockRoot)
	if err != nil {
		return nil, err
	}
	if ok {
		log.WithFields(logrus.Fields{
			"0slot": cachedInfo.state.Slot(),
			"root":  fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: from epoch boundary cache")
		return cachedInfo.state, nil
	}

	// check the state exists in replay cache
	resState, err := s.loadStateCache.Get(ctx, blockRoot)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Error("Load state by root: replay cache get failed")
		return nil, err
	}
	if resState != nil {
		log.WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: from replay cache get")
		return resState, nil
	}

	// Short cut if the cachedState is already in the DB.
	if s.beaconDB.HasState(ctx, blockRoot) {
		log.WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: from DB")
		return s.beaconDB.State(ctx, blockRoot)
	}

	// if same request is already in progress - waite result
	resState, err = s.loadStateCache.GetWhenReady(ctx, blockRoot)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Error("Load state by root: replay cache wait failed")
		return nil, err
	}
	if resState != nil {
		log.WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Info("Load state by root: from replay wait cache")
		return resState, nil
	}
	if err = s.loadStateCache.MarkInProgress(blockRoot); err != nil {
		if errors.Is(err, cache.ErrAlreadyInProgress) {
			resState, err = s.loadStateCache.GetWhenReady(ctx, blockRoot)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"root": fmt.Sprintf("%#x", blockRoot),
				}).Error("Load state by root: replay cache already progress")
				return nil, err
			}
			if resState == nil || resState.IsNil() {
				err = errors.New("replay cache was in progress and resolved nil")
				log.WithError(err).WithFields(logrus.Fields{
					"root": fmt.Sprintf("%#x", blockRoot),
				}).Error("Load state by root: replay cache resolved nil")
				return nil, err
			}
			log.WithFields(logrus.Fields{
				"root": fmt.Sprintf("%#x", blockRoot),
			}).Info("Load state by root: from replay cache 2")
			return resState, nil
		}
		log.WithError(err).WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Error("Load state by root: replay cache start failed")
		return nil, err
	}
	defer func() {
		if err := s.loadStateCache.MarkNotInProgress(blockRoot); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"root": fmt.Sprintf("%#x", blockRoot),
			}).Error("Load state by root: replay cache failed to mark in progress")
		}
	}()

	log.WithFields(logrus.Fields{
		"root": fmt.Sprintf("%#x", blockRoot),
	}).Info("Load state by root: replay blocks start")

	summary, err := s.stateSummary(ctx, blockRoot)
	if err != nil {
		return nil, errors.Wrap(err, "could not get state summary")
	}
	targetSlot := summary.Slot

	// Since the requested state is not in caches, start replaying using the last available ancestor state which is
	// retrieved using input block's parent root.
	startState, err := s.LastAncestorState(ctx, blockRoot)
	if err != nil {
		return nil, errors.Wrap(err, "could not get ancestor state")
	}
	if startState == nil || startState.IsNil() {
		return nil, errUnknownBoundaryState
	}

	// Return state early if we are retrieving it from our finalized state cache.
	if startState.Slot() == targetSlot {
		return startState, nil
	}

	blks, err := s.LoadBlocks(ctx, startState.Slot()+1, targetSlot, bytesutil.ToBytes32(summary.Root))
	if err != nil {
		return nil, errors.Wrap(err, "could not load blocks for hot state using root")
	}

	replayBlockCount.Observe(float64(len(blks)))

	resState, err = s.ReplayBlocks(ctx, startState, blks, targetSlot)
	if err != nil {
		return resState, err
	}
	// cache result of the ReplayBlocks method
	if err = s.loadStateCache.Put(ctx, blockRoot, resState); err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"root": fmt.Sprintf("%#x", blockRoot),
		}).Error("Load state by root: replay cache put failed")
	}
	return resState, nil
}

// This returns the highest available ancestor state of the input block root.
// It recursively look up block's parent until a corresponding state of the block root
// is found in the caches or DB.
//
// There's three ways to derive block parent state:
// 1.) block parent state is the last finalized state
// 2.) block parent state is the epoch boundary state and exists in epoch boundary cache.
// 3.) block parent state is in DB.
func (s *State) LastAncestorState(ctx context.Context, root [32]byte) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "stateGen.LastAncestorState")
	defer span.End()

	if s.isFinalizedRoot(root) && s.finalizedState() != nil {
		return s.finalizedState(), nil
	}

	b, err := s.beaconDB.Block(ctx, root)
	if err != nil {
		return nil, err
	}
	if err := helpers.BeaconBlockIsNil(b); err != nil {
		return nil, err
	}

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Is the state a genesis state.
		parentRoot := bytesutil.ToBytes32(b.Block().ParentRoot())
		if parentRoot == params.BeaconConfig().ZeroHash {
			return s.beaconDB.GenesisState(ctx)
		}

		// return an error if slot hasn't been covered by checkpoint sync backfill
		ps := b.Block().Slot() - 1
		if !s.slotAvailable(ps) {
			return nil, errors.Wrapf(ErrNoDataForSlot, "slot %d not in db due to checkpoint sync", ps)
		}
		// Does the state exist in the hot state cache.
		if s.hotStateCache.has(parentRoot) {
			return s.hotStateCache.get(parentRoot), nil
		}

		if s.syncStateCache.has(parentRoot) {
			return s.syncStateCache.get(parentRoot), nil
		}

		// Does the state exist in finalized info cache.
		if s.isFinalizedRoot(parentRoot) {
			return s.finalizedState(), nil
		}

		// Does the state exist in epoch boundary cache.
		cachedInfo, ok, err := s.epochBoundaryStateCache.getByRoot(parentRoot)
		if err != nil {
			return nil, err
		}
		if ok {
			return cachedInfo.state, nil
		}

		// Does the state exists in DB.
		if s.beaconDB.HasState(ctx, parentRoot) {
			return s.beaconDB.State(ctx, parentRoot)
		}
		b, err = s.beaconDB.Block(ctx, parentRoot)
		if err != nil {
			return nil, err
		}
		if b == nil || b.IsNil() {
			return nil, errUnknownBlock
		}
	}
}

func (s *State) CombinedCache() *CombinedCache {
	getters := make([]CachedGetter, 0)
	if s.hotStateCache != nil {
		getters = append(getters, s.hotStateCache)
	}
	if s.epochBoundaryStateCache != nil {
		getters = append(getters, s.epochBoundaryStateCache)
	}
	return &CombinedCache{getters: getters}
}

func (s *State) slotAvailable(slot types.Slot) bool {
	// default to assuming node was initialized from genesis - backfill only needs to be specified for checkpoint sync
	if s.backfillStatus == nil {
		return true
	}
	return s.backfillStatus.SlotCovered(slot)
}
