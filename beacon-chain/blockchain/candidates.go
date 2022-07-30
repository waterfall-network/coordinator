package blockchain

import (
	"github.com/waterfall-foundation/gwat/common"
	"github.com/waterfall-foundation/gwat/dag/finalizer"
)

// setCandidates cashes current candidates.
func (s *Service) setCandidates(candidates finalizer.NrHashMap) {
	//contains at least last known candidate
	if len(candidates) == 0 || candidates.HasGap() {
		hash, nr := s.GetLastCandidate()
		if hash != nil {
			s.candidates = &finalizer.NrHashMap{nr: hash}
		}
		return
	}
	s.candidates = candidates.Copy()
}

// GetCandidates returns current candidates.
func (s *Service) GetCandidates() finalizer.NrHashMap {
	if s.candidates == nil || s.candidates.Copy() == nil || len(*s.candidates.Copy()) == 0 {
		hash, nr := s.GetLastCandidate()
		if hash == nil || nr == 0 {
			return finalizer.NrHashMap{}
		}
		return finalizer.NrHashMap{nr: hash}
	}
	return *s.candidates.Copy()
}

func (s *Service) GetLastCandidate() (*common.Hash, uint64) {
	pNhm := s.candidates
	if pNhm == nil || pNhm.Copy() == nil || len(*pNhm.Copy()) == 0 {
		pNhm.SetBytes(s.cfg.ChainStartFetcher.ChainStartEth1Data().Candidates)
		if pNhm != nil && len(*pNhm) > 0 {
			maxNr := *pNhm.GetMaxNr()
			return (*pNhm)[maxNr], maxNr
		}
		return nil, 0
	}
	maxNr := pNhm.GetMaxNr()
	if maxNr == nil {
		return nil, 0
	}
	return (*pNhm)[*maxNr], *maxNr
}
