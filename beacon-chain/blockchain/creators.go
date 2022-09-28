package blockchain

import (
	"sync"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/time/slots"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

// This defines the current chain service's view of creators.
type creatorsAssignment struct {
	assignment map[types.Slot][]gwatCommon.Address // creators' assignment by slot.
	lock       sync.RWMutex
}

// GetCurrentCreators returns creators assignments for current slot.
func (s *Service) GetCurrentCreators() ([]gwatCommon.Address, error) {
	s.creators.lock.RLock()
	defer s.creators.lock.RUnlock()
	slot := s.CurrentSlot()
	// retrieve creators assignments from cache
	if s.creators.assignment != nil && s.creators.assignment[slot] != nil {
		return s.creators.assignment[slot], nil
	}

	// calculate creators assignments
	ctx := s.ctx
	state := s.headState(ctx)
	epoch := slots.ToEpoch(slot)
	creatorsAssig, err := helpers.CalcCreatorsAssignments(ctx, state, epoch)
	if err != nil {
		return []gwatCommon.Address{}, err
	}
	s.creators.assignment = creatorsAssig
	return s.creators.assignment[slot], nil
}
