package blockchain

import (
	"sync"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

type spineData struct {
	candidates    gwatCommon.HashArray
	actual        bool
	lastValidRoot []byte
	lastValidSlot types.Slot
	sync.RWMutex
}

// SetValidatedBlockInfo cashes info of the latest success validated block.
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

// setCacheCandidates cashes current candidates.
func (s *Service) setCacheCandidates(candidates gwatCommon.HashArray) {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	s.spineData.candidates = candidates.Copy()
}

// GetCacheCandidates returns current candidates.
func (s *Service) GetCacheCandidates() gwatCommon.HashArray {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	return s.spineData.candidates.Copy()
}

// setCacheCandidates cashes current candidates.
func (s *Service) setCandidatesActual(actual bool) {
	s.spineData.RLock()
	defer s.spineData.RUnlock()
	s.spineData.actual = actual
}

// GetCacheCandidates returns current candidates.
func (s *Service) AreCandidatesActual() bool {
	s.spineData.RLock()
	defer s.spineData.RUnlock()
	return s.spineData.actual
}

// ValidateBlockCandidates validate new block candidates.
func (s *Service) ValidateBlockCandidates(block block.BeaconBlock) (bool, error) {
	blCandidates := gwatCommon.HashArrayFromBytes(block.Body().Eth1Data().Candidates)

	log.WithFields(logrus.Fields{
		"slot":            block.Slot(),
		"blockCandidates": blCandidates,
	}).Info("**** Blocks Candidates Validation v2 ****")

	if len(blCandidates) == 0 {
		return true, nil
	}
	return s.cfg.ExecutionEngineCaller.ExecutionDagValidateSpines(s.ctx, blCandidates)
}
