package stategen

import (
	"context"
	"fmt"
	"math"

	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	"go.opencensus.io/trace"
)

// SaveState saves the state in the cache and/or DB.
func (s *State) SaveState(ctx context.Context, root [32]byte, st state.BeaconState) error {
	ctx, span := trace.StartSpan(ctx, "stateGen.SaveState")
	defer span.End()
	return s.saveStateByRoot(ctx, root, st)
}

// ForceCheckpoint initiates a cold state save of the given state. This method does not update the
// "last archived state" but simply saves the specified state from the root argument into the DB.
func (s *State) ForceCheckpoint(ctx context.Context, root []byte) error {
	ctx, span := trace.StartSpan(ctx, "stateGen.ForceCheckpoint")
	defer span.End()

	root32 := bytesutil.ToBytes32(root)
	// Before the first finalized check point, the finalized root is zero hash.
	// Return early if there hasn't been a finalized check point.
	if root32 == params.BeaconConfig().ZeroHash {
		return nil
	}

	fs, err := s.loadStateByRoot(ctx, root32)
	if err != nil {
		return err
	}

	return s.beaconDB.SaveState(ctx, fs, root32)
}

// This saves a post beacon state. On the epoch boundary,
// it saves a full state. On an intermediate slot, it saves a back pointer to the
// nearest epoch boundary state.
func (s *State) saveStateByRoot(ctx context.Context, blockRoot [32]byte, st state.BeaconState) error {
	ctx, span := trace.StartSpan(ctx, "stateGen.saveStateByRoot")
	defer span.End()

	// Duration can't be 0 to prevent panic for division.
	duration := uint64(math.Max(float64(s.saveHotStateDB.duration), 1))

	s.saveHotStateDB.lock.Lock()
	if s.saveHotStateDB.enabled && st.Slot().Mod(duration) == 0 {
		if err := s.beaconDB.SaveState(ctx, st, blockRoot); err != nil {
			s.saveHotStateDB.lock.Unlock()
			log.WithError(err).WithFields(logrus.Fields{
				"epoch": slots.ToEpoch(st.Slot()),
				"slot":  fmt.Sprintf("%d", st.Slot()),
				"root":  fmt.Sprintf("%#x", blockRoot),
			}).Error("Save state by root: save to db failed")
			return err
		}
		s.saveHotStateDB.savedStateRoots = append(s.saveHotStateDB.savedStateRoots, blockRoot)

		log.WithFields(logrus.Fields{
			"slot":                   st.Slot(),
			"totalHotStateSavedInDB": len(s.saveHotStateDB.savedStateRoots),
			"root":                   fmt.Sprintf("%#x", blockRoot),
		}).Info("Save state by root: save to db success")
	}
	s.saveHotStateDB.lock.Unlock()

	// If the hot state is already in cache, one can be sure the state was processed and in the DB.
	if s.hotStateCache.has(blockRoot) {
		return nil
	}

	// Only on an epoch boundary slot, saves epoch boundary state in epoch boundary root state cache.
	if slots.IsEpochStart(st.Slot()) {
		if err := s.epochBoundaryStateCache.put(blockRoot, st); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"epoch": slots.ToEpoch(st.Slot()),
				"slot":  fmt.Sprintf("%d", st.Slot()),
				"root":  fmt.Sprintf("%#x", blockRoot),
			}).Error("Save state by root: add epoch boundary cache failed")
			return err
		}
		log.WithFields(logrus.Fields{
			"0epoch": slots.ToEpoch(st.Slot()),
			"1slot":  fmt.Sprintf("%d", st.Slot()),
			"root":   fmt.Sprintf("%#x", blockRoot),
		}).Debug("Save state by root: add epoch boundary cache success")
	} else {
		// Always check that the correct epoch boundary states have been saved
		// for the current epoch.
		epochStart, err := slots.EpochStart(slots.ToEpoch(st.Slot()))
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"slot": epochStart,
				"root": fmt.Sprintf("%#x", blockRoot),
			}).Error("Save state by root: epoch start failed")
			return err
		}
		bRoot, err := helpers.BlockRootAtSlot(st, epochStart)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"slot": epochStart,
				"root": fmt.Sprintf("%#x", bRoot),
			}).Error("Save state by root: check block at slot failed")
			return err
		}
		_, ok, err := s.epochBoundaryStateCache.getByRoot([32]byte(bRoot))
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"slot": epochStart,
				"root": fmt.Sprintf("%#x", bRoot),
			}).Error("Save state by root: check epoch boundary cache failed")
			return err
		}
		// We would only recover the boundary states under this condition:
		//
		// 1) Would indicate that the epoch boundary was skipped due to a missed slot, we
		// then recover by saving the state at that particular slot here.
		if !ok {
			// Only recover the state if it is in our hot state cache, otherwise we
			// simply skip this step.
			if s.hotStateCache.has([32]byte(bRoot)) {
				log.WithFields(logrus.Fields{
					"slot": epochStart,
					"root": fmt.Sprintf("%#x", bRoot),
				}).Debug("Save state by root: recovering state")

				hState := s.hotStateCache.get([32]byte(bRoot))
				if err := s.epochBoundaryStateCache.put([32]byte(bRoot), hState); err != nil {
					log.WithError(err).WithFields(logrus.Fields{
						"slot": epochStart,
						"root": fmt.Sprintf("%#x", bRoot),
					}).Error("Save state by root: recovering state failed")
					return err
				}
			}
		}
		log.WithFields(logrus.Fields{
			"0epoch":        slots.ToEpoch(st.Slot()),
			"1slot":         fmt.Sprintf("%d", st.Slot()),
			"2cachedBefore": ok,
			"root":          fmt.Sprintf("%#x", blockRoot),
		}).Debug("Save state by root: add epoch boundary cache success")
	}

	// On an intermediate slots, save state summary.
	if err := s.beaconDB.SaveStateSummary(ctx, &ethpb.StateSummary{
		Slot: st.Slot(),
		Root: blockRoot[:],
	}); err != nil {
		return err
	}

	// Store the copied state in the hot state cache.
	s.hotStateCache.put(blockRoot, st.Copy())

	return nil
}

// EnableSaveHotStateToDB enters the mode that saves hot beacon state to the DB.
// This usually gets triggered when there's long duration since finality.
func (s *State) EnableSaveHotStateToDB(_ context.Context) {
	s.saveHotStateDB.lock.Lock()
	defer s.saveHotStateDB.lock.Unlock()
	if s.saveHotStateDB.enabled {
		return
	}

	s.saveHotStateDB.enabled = true

	log.WithFields(logrus.Fields{
		"enabled":       s.saveHotStateDB.enabled,
		"slotsInterval": s.saveHotStateDB.duration,
	}).Warn("Entering mode to save hot states in DB")
}

// DisableSaveHotStateToDB exits the mode that saves beacon state to DB for the hot states.
// This usually gets triggered once there's finality after long duration since finality.
func (s *State) DisableSaveHotStateToDB(ctx context.Context) error {
	s.saveHotStateDB.lock.Lock()
	defer s.saveHotStateDB.lock.Unlock()
	if !s.saveHotStateDB.enabled {
		return nil
	}

	log.WithFields(logrus.Fields{
		"enabled":          s.saveHotStateDB.enabled,
		"deletedHotStates": len(s.saveHotStateDB.savedStateRoots),
	}).Warn("Exiting mode to save hot states in DB")

	// Delete previous saved states in DB as we are turning this mode off.
	s.saveHotStateDB.enabled = false
	for _, r := range s.saveHotStateDB.savedStateRoots {
		if err := s.beaconDB.DeleteState(ctx, r); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"root": fmt.Sprintf("%#x", r),
			}).Warn("Exiting mode to save hot states in DB: delete stale state ignored")
		}
	}
	//if err := s.beaconDB.DeleteStates(ctx, s.saveHotStateDB.savedStateRoots); err != nil {
	//	return err
	//}
	s.saveHotStateDB.savedStateRoots = nil

	return nil
}

// AddSyncStateCache caches state to sync cache.
func (s *State) AddSyncStateCache(blockRoot [32]byte, st state.BeaconState) error {
	s.syncStateCache.put(blockRoot, st.Copy())
	return nil
}

// AddSyncStateCache caches state to sync cache.
func (s *State) RemoveSyncStateCache(blockRoot [32]byte) {
	s.syncStateCache.delete(blockRoot)
}

// AddSyncStateCache caches state to sync cache.
func (s *State) PurgeSyncStateCache() {
	s.syncStateCache.purge()
}
