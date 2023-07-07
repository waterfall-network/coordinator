package blockchain

import (
	"bytes"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	types "github.com/prysmaticlabs/eth2-types"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	gwatTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
)

type spineData struct {
	optSpines   *lru.Cache
	optSpinesMu sync.RWMutex

	lastValidRoot []byte
	lastValidSlot types.Slot
	//finalizedSpines gwatCommon.HashArray  //successfully finalized spines from checkpoint
	gwatCheckpoint *gwatTypes.Checkpoint //cache for current finalization request checkpoint param
	coordState     *gwatTypes.Checkpoint //cache for current gwat coordinated state
	sync.RWMutex
}

// setCacheOptimisticSpines cashes current optSpines.
func (s *Service) setCacheOptimisticSpines(baseSpine gwatCommon.Hash, optSpines []gwatCommon.HashArray) {
	s.spineData.optSpinesMu.RLock()
	defer s.spineData.optSpinesMu.RUnlock()
	if s.spineData.optSpines == nil {
		var err error
		s.spineData.optSpines, err = lru.New(8)
		if err != nil {
			log.WithError(err).Error("create optSpines failed")
		}
	}
	s.spineData.optSpines.Remove(baseSpine)
	s.spineData.optSpines.Add(baseSpine, optSpines)
}

// GetCacheOptimisticSpines returns current optSpines.
func (s *Service) GetCacheOptimisticSpines(baseSpine gwatCommon.Hash) []gwatCommon.HashArray {
	s.spineData.optSpinesMu.RLock()
	defer s.spineData.optSpinesMu.RUnlock()

	if s.spineData.optSpines == nil {
		return []gwatCommon.HashArray{}
	}

	data, ok := s.spineData.optSpines.Get(baseSpine)
	if !ok {
		return []gwatCommon.HashArray{}
	}
	optSpines, ok := data.([]gwatCommon.HashArray)
	if !ok {
		return []gwatCommon.HashArray{}
	}
	if optSpines != nil {
		return optSpines
	}
	return []gwatCommon.HashArray{}
}

// SetValidatedBlockInfo caches info of the latest success validated block.
func (s *Service) SetValidatedBlockInfo(lastValidRoot []byte, lastValidSlot types.Slot) {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	s.spineData.lastValidRoot = lastValidRoot
	s.spineData.lastValidSlot = lastValidSlot
}

// GetValidatedBlockInfo returns info of the latest success validated block.
func (s *Service) GetValidatedBlockInfo() ([]byte, types.Slot) {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	return s.spineData.lastValidRoot, s.spineData.lastValidSlot
}

//// SetFinalizedSpinesCheckpoint set spine hash of checkpoint
//func (s *Service) SetFinalizedSpinesCheckpoint(cpSpine gwatCommon.Hash) {
//	s.spineData.RLock()
//	defer s.spineData.RUnlock()
//
//	finalizedSpines := make(gwatCommon.HashArray, 0, 4*params.BeaconConfig().SlotsPerEpoch)
//	isCpFound := false
//	for _, h := range s.spineData.finalizedSpines {
//		if h == cpSpine || isCpFound {
//			isCpFound = true
//			finalizedSpines = append(finalizedSpines, h)
//		}
//	}
//	if len(finalizedSpines) == 0 {
//		finalizedSpines = append(finalizedSpines, cpSpine)
//	}
//	s.spineData.finalizedSpines = finalizedSpines.Uniq()
//}

//// AddFinalizedSpines append finalized spines to cache
//func (s *Service) AddFinalizedSpines(finSpines gwatCommon.HashArray) {
//	s.spineData.RLock()
//	defer s.spineData.RUnlock()
//
//	if len(finSpines) == 0 {
//		return
//	}
//	firstSpine := finSpines[0]
//	finalizedSpines := make(gwatCommon.HashArray, 0, 4*params.BeaconConfig().SlotsPerEpoch)
//	for _, h := range s.spineData.finalizedSpines {
//		if h == firstSpine {
//			break
//		}
//		finalizedSpines = append(finalizedSpines, h)
//	}
//	finalizedSpines = append(finalizedSpines, finSpines...)
//	s.spineData.finalizedSpines = finalizedSpines.Uniq()
//}

//func (s *Service) SetFinalizedSpinesHead(headSpine gwatCommon.Hash) {
//	s.AddFinalizedSpines(gwatCommon.HashArray{headSpine})
//}

//func (s *Service) GetFinalizedSpines() gwatCommon.HashArray {
//	s.spineData.RLock()
//	defer s.spineData.RUnlock()
//	return s.spineData.finalizedSpines.Copy()
//}

//func (s *Service) ResetFinalizedSpines() {
//	s.spineData.RLock()
//	defer s.spineData.RUnlock()
//
//	finalizedSpines := make(gwatCommon.HashArray, 0, 4*params.BeaconConfig().SlotsPerEpoch)
//	s.spineData.finalizedSpines = finalizedSpines
//}

// CacheGwatCheckpoint caches the current gwat checkpoint.
func (s *Service) CacheGwatCheckpoint(gwatCheckpoint *gwatTypes.Checkpoint) {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	s.spineData.gwatCheckpoint = gwatCheckpoint
}

// GetCachedGwatCheckpoint returns the currently cached gwat checkpoint.
func (s *Service) GetCachedGwatCheckpoint(cpRoot []byte) *gwatTypes.Checkpoint {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	cp := s.spineData.gwatCheckpoint
	if cp != nil && bytes.Equal(cp.Root.Bytes(), cpRoot) {
		return s.spineData.gwatCheckpoint
	}
	return nil
}

// CacheGwatCoordinatedState caches the current gwat coordinated state.
func (s *Service) CacheGwatCoordinatedState(coordState *gwatTypes.Checkpoint) {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	s.spineData.coordState = coordState
}

// GetCachedGwatCoordinatedState returns the currently cached gwat coordinated state.
func (s *Service) GetCachedGwatCoordinatedState() *gwatTypes.Checkpoint {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	return s.spineData.coordState
}
