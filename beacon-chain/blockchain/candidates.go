package blockchain

import (
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

// setCacheCandidates cashes current candidates.
func (s *Service) setCacheCandidates(candidates gwatCommon.HashArray) {
	s.candidates = candidates.Copy()
}

// GetCandidates returns current candidates.
func (s *Service) GetCacheCandidates() gwatCommon.HashArray {
	return s.candidates.Copy()
}
