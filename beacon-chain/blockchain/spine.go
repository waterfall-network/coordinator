package blockchain

import (
	"sync"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

type spineData struct {
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

// ValidateBlockCandidates validate new block candidates.
func (s *Service) ValidateBlockCandidates(block block.BeaconBlock) (bool, error) {
	blCandidates := gwatCommon.HashArrayFromBytes(block.Body().Eth1Data().Candidates)
	if len(blCandidates) == 0 {
		return true, nil
	}
	return s.cfg.ExecutionEngineCaller.ExecutionDagValidateSpines(s.ctx, blCandidates)
}
