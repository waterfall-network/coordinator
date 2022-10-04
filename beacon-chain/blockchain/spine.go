package blockchain

import (
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/block"
	"sort"
	"sync"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/encoding/bytesutil"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

const RequiredVotesPct = 66

type mapVoting map[gwatCommon.Hash]int
type mapPriority map[int]gwatCommon.HashArray
type mapCandidates map[gwatCommon.Hash]gwatCommon.HashArray

type spineData struct {
	candidates   gwatCommon.HashArray
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

// setCacheFinalisation cashes current spines for finalization.
func (s *Service) setCacheFinalisation(spines gwatCommon.HashArray) {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	if len(spines) > 0 {
		s.spineData.finalization = spines.Copy()
	}
}

// GetCacheFinalization returns current spines to finalization.
func (s *Service) GetCacheFinalization() gwatCommon.HashArray {
	s.spineData.RLock()
	defer s.spineData.RUnlock()

	return s.spineData.finalization.Copy()
}

// GetLastFinSpine returns current candidates.
func (s *Service) GetLastFinSpine() (gwatCommon.Hash, types.Slot, error) {
	s.spineData.RLock()
	defer s.spineData.RUnlock()
	if len(s.spineData.finalization) == 0 {
		// retrieve last finalized header
		header, err := s.cfg.ExecutionEngineCaller.GetHeaderByNumber(s.ctx, nil)
		if err != nil || header == nil {
			log.WithError(err).Error(errRetrievingSpineFailed.Error())
			return gwatCommon.Hash{}, 0, errRetrievingSpineFailed
		}
		//set finalization data
		s.spineData.finalization = gwatCommon.HashArray{header.Hash()}
		s.spineData.lastFinHash = header.Hash()
		s.spineData.lastFinSlot = types.Slot(header.Slot)
		return s.spineData.lastFinHash, s.spineData.lastFinSlot, nil
	}
	lastHash := s.spineData.finalization[len(s.spineData.finalization)-1]
	// if finalization is unchanged used cache
	if s.spineData.lastFinHash == lastHash {
		return s.spineData.lastFinHash, s.spineData.lastFinSlot, nil
	}
	// retrieve last finalized header
	header, err := s.cfg.ExecutionEngineCaller.GetHeaderByNumber(s.ctx, nil)
	if err != nil || header == nil {
		log.WithError(err).Error(errRetrievingSpineFailed.Error())
		return gwatCommon.Hash{}, 0, errRetrievingSpineFailed
	}
	// if Finalization fully finalized
	//  - cutoff it to last spine
	if header.Hash() == lastHash {
		s.spineData.finalization = gwatCommon.HashArray{header.Hash()}
	}
	s.spineData.lastFinHash = header.Hash()
	s.spineData.lastFinSlot = types.Slot(header.Slot)
	return s.spineData.lastFinHash, s.spineData.lastFinSlot, nil
}

func getVoteNeeded() int {
	return 6
}

func (s *Service) CalculateFinalizationSpinesByBlockRoot(blockRoot [32]byte) (gwatCommon.HashArray, error) {
	var (
		candidatesList = []gwatCommon.HashArray{}
		tabPriority    = mapPriority{}
		tabVoting      = mapVoting{}
		tabCandidates  = mapCandidates{}
		votesNeeded    = getVoteNeeded()
	)

	lastSpineHash, lastSpineSlot, err := s.GetLastFinSpine()
	if err != nil || lastSpineHash == (gwatCommon.Hash{}) {
		return gwatCommon.HashArray{}, err
	}
	// collect all beacon blocks up to last finalized gwat block
	currRoot := blockRoot
	for {
		if currRoot == params.BeaconConfig().ZeroHash {
			return gwatCommon.HashArray{}, nil
		}
		block, err := s.cfg.BeaconDB.Block(s.ctx, currRoot)
		if err != nil {
			return gwatCommon.HashArray{}, err
		}
		// if reach finalized slot
		if block.Block().Slot() <= lastSpineSlot {
			break
		}

		candidates := gwatCommon.HashArrayFromBytes(block.Block().Body().Eth1Data().Candidates)

		if !candidates.IsUniq() {
			log.WithField("candidates", candidates).Warn("skip bad candidates: is not uniq")
			continue
		}
		//exclude finalized spines
		fullLen := len(candidates)
		if i := candidates.IndexOf(lastSpineHash); i >= 0 {
			candidates = candidates[i+1:]
		}
		// if all current candidates handled
		if len(candidates) == 0 && fullLen > len(candidates) {
			break
		}
		candidatesList = append(candidatesList, candidates)
		// reduction of sequence up to single item
		for i := len(candidates) - 1; i > 0; i-- {
			reduction := candidates[:i]
			candidatesList = append(candidatesList, reduction)
		}
		//set next block root
		currRoot = bytesutil.ToBytes32(block.Block().ParentRoot())
	}

	//calculate voting params
	for i, candidates := range candidatesList {
		restList := candidatesList[i+1:]
		if !candidates.IsUniq() {
			log.WithField("candidates", candidates).Warn("bad candidates: is not uniq")
			continue
		}
		for _, rc := range restList {
			intersect := candidates.SequenceIntersection(rc)
			key := intersect.Key()
			tabCandidates[key] = intersect
			tabPriority[len(intersect)] = append(tabPriority[len(intersect)], key).Uniq()
			tabVoting[key]++
		}
	}

	//sort by priority
	priorities := []int{}
	for p, _ := range tabPriority {
		priorities = append(priorities, p)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(priorities)))

	// find voting result
	resKey := gwatCommon.Hash{}
	for _, p := range priorities {
		// select by max number of vote which satisfies the condition
		// of min required number of votes
		maxVotes := 0
		for _, key := range tabPriority[p] {
			votes := tabVoting[key]
			if votes >= votesNeeded && votes > maxVotes {
				resKey = key
			}
		}
		if resKey != (gwatCommon.Hash{}) {
			break
		}
	}

	log.WithFields(logrus.Fields{
		"lastFinSlot": lastSpineSlot,
		"lastFinHash": lastSpineHash.Hex(),
		"spines":      tabCandidates[resKey],
	}).Info("Calculation of finalization sequence")

	if resKey == (gwatCommon.Hash{}) {
		return gwatCommon.HashArray{}, nil
	}
	return tabCandidates[resKey], nil
}

// todo deprecated, don't use
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
