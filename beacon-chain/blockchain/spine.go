package blockchain

import (
	"sync"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

//type mapVoting map[gwatCommon.Hash]int
//type mapPriority map[int]gwatCommon.HashArray
//type mapCandidates map[gwatCommon.Hash]gwatCommon.HashArray

type spineData struct {
	candidates gwatCommon.HashArray
	actual     bool

	finalization gwatCommon.HashArray
	lastFinHash  gwatCommon.Hash
	lastFinSlot  types.Slot
	sync.RWMutex
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

//// ValidateBlockCandidates validate new block candidates.
//// todo depracated
//func (s *Service) ValidateBlockCandidates(block block.BeaconBlock) (bool, error) {
//	slot := block.Slot()
//	blCandidates := gwatCommon.HashArrayFromBytes(block.Body().Eth1Data().Candidates)
//	lenBlC := len(blCandidates)
//	if lenBlC == 0 {
//		return true, nil
//	}
//	slotCandidates, err := s.cfg.ExecutionEngineCaller.ExecutionDagGetCandidates(s.ctx, slot)
//	if err != nil {
//		log.WithError(err).WithFields(logrus.Fields{
//			"slot":            slot,
//			"blockCandidates": blCandidates,
//			"gwatCandidates":  slotCandidates,
//		}).Warn("**** Blocks Candidates Validation: err ****")
//		return false, err
//	}
//	startIx := slotCandidates.IndexOf(blCandidates[0])
//	endIx := slotCandidates.IndexOf(blCandidates[lenBlC-1])
//	if startIx < 0 || endIx < 0 {
//		log.WithFields(logrus.Fields{
//			"slot":            slot,
//			"condition":       "startIx < 0 || endIx < 0",
//			"startIx":         startIx,
//			"endIx":           endIx,
//			"blockCandidates": blCandidates,
//			"gwatCandidates":  slotCandidates,
//		}).Warn("**** Blocks Candidates Validation: failed ****")
//
//		return false, nil
//	}
//	validCandidates := slotCandidates[startIx : endIx+1]
//	isValid := validCandidates.IsEqualTo(blCandidates)
//	if !isValid {
//		log.WithFields(logrus.Fields{
//			"isValid":         isValid,
//			"slot":            slot,
//			"blockCandidates": blCandidates,
//			"gwatCandidates":  slotCandidates,
//		}).Warn("**** Blocks Candidates Validation: failed ****")
//	}
//	return isValid, nil
//}

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
