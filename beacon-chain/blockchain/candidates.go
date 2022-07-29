package blockchain

import (
	"github.com/waterfall-foundation/gwat/dag/finalizer"
)

// setCandidates cashes current candidates.
func (s *Service) setCandidates(candidates finalizer.NrHashMap) {
	s.candidates = candidates.Copy()
}

// GetCandidates returns current candidates.
func (s *Service) GetCandidates() finalizer.NrHashMap {
	if s.candidates == nil || s.candidates.Copy() == nil {
		return finalizer.NrHashMap{}
	}
	return *s.candidates.Copy()
}
