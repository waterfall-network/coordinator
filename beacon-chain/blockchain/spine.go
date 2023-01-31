package blockchain

import (
	"sync"

	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/block"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

type spineData struct {
	lastValidRoot   []byte
	lastValidSlot   types.Slot
	finalizedSpines gwatCommon.HashArray //successfully finalized spines from checkpoint
	sync.RWMutex
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

// ValidateBlockCandidates validate new block candidates.
func (s *Service) ValidateBlockCandidates(block block.BeaconBlock) (bool, error) {
	blCandidates := gwatCommon.HashArrayFromBytes(block.Body().Eth1Data().Candidates)
	if len(blCandidates) == 0 {
		return true, nil
	}
	return s.cfg.ExecutionEngineCaller.ExecutionDagValidateSpines(s.ctx, blCandidates)
}

// SetFinalizedSpinesCheckpoint set spine hash of checkpoint
func (s *Service) SetFinalizedSpinesCheckpoint(cpSpine gwatCommon.Hash) {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	finalizedSpines := make(gwatCommon.HashArray, 0, 4*params.BeaconConfig().SlotsPerEpoch)
	isCpFound := false
	for _, h := range s.spineData.finalizedSpines {
		if h == cpSpine || isCpFound {
			isCpFound = true
			finalizedSpines = append(finalizedSpines, h)
		}
	}
	if len(finalizedSpines) == 0 {
		finalizedSpines = append(finalizedSpines, cpSpine)
	}
	s.spineData.finalizedSpines = finalizedSpines.Uniq()
}

// AddFinalizedSpines append finalized spines to cache
func (s *Service) AddFinalizedSpines(finSpines gwatCommon.HashArray) {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	if len(finSpines) == 0 {
		return
	}
	firstSpine := finSpines[0]
	finalizedSpines := make(gwatCommon.HashArray, 0, 4*params.BeaconConfig().SlotsPerEpoch)
	for _, h := range s.spineData.finalizedSpines {
		if h == firstSpine {
			break
		}
		finalizedSpines = append(finalizedSpines, h)
	}
	finalizedSpines = append(finalizedSpines, finSpines...)
	s.spineData.finalizedSpines = finalizedSpines.Uniq()
}

func (s *Service) SetFinalizedSpinesHead(headSpine gwatCommon.Hash) {
	s.AddFinalizedSpines(gwatCommon.HashArray{headSpine})
}

func (s *Service) GetFinalizedSpines() gwatCommon.HashArray {
	s.spineData.RLock()
	defer s.spineData.RUnlock()
	return s.spineData.finalizedSpines.Copy()
}

func (s *Service) ResetFinalizedSpines() {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	finalizedSpines := make(gwatCommon.HashArray, 0, 4*params.BeaconConfig().SlotsPerEpoch)
	s.spineData.finalizedSpines = finalizedSpines
}
