package blockchain

import (
	"github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

// ValidateBlockCandidates validate new block candidates.
func (s *Service) ValidateBlockCandidates(block block.BeaconBlock) (bool, error) {
	slot := block.Slot()
	blCandidates := gwatCommon.HashArrayFromBytes(block.Body().Eth1Data().Candidates)
	lenBlC := len(blCandidates)
	if lenBlC == 0 {
		return true, nil
	}
	slotCandidates, err := s.cfg.ExecutionEngineCaller.ExecutionDagGetCandidates(s.ctx, slot)
	if err != nil {
		return false, err
	}
	startIx := slotCandidates.IndexOf(blCandidates[0])
	endIx := slotCandidates.IndexOf(blCandidates[lenBlC-1])
	if startIx < 0 || endIx < 0 {
		return false, nil
	}
	validCandidates := slotCandidates[startIx : endIx+1]
	return validCandidates.IsEqualTo(blCandidates), nil
}
