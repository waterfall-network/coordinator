package blockchain

import (
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

// setCandidates cashes current candidates.
func (s *Service) setCandidates(candidates gwatCommon.HashArray) {
	s.candidates = candidates.Copy()
}

// GetCandidates returns current candidates.
func (s *Service) GetCandidates() gwatCommon.HashArray {
	return s.candidates.Copy()
}
