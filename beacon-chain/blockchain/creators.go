package blockchain

import (
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/config/params"
	"github.com/prysmaticlabs/prysm/time/slots"
	"github.com/waterfall-foundation/gwat/common"
	"sync"
)

// This defines the current chain service's view of creators.
type creatorsAssignment struct {
	assignment map[types.Slot][]common.Address // creators' assignment by slot.
	lock       sync.RWMutex
}

// GetCurrentCreators returns creators assignments for current slot.
func (s *Service) GetCurrentCreators() ([]common.Address, error) {
	s.creators.lock.RLock()
	defer s.creators.lock.RUnlock()
	slot := s.HeadSlot() + 1
	// retrieve creators assignments from cache
	if s.creators.assignment != nil && s.creators.assignment[slot] != nil {
		return s.creators.assignment[slot], nil
	}

	// calculate creators assignments
	creatorsAssig := make(map[types.Slot][]common.Address, params.BeaconConfig().SlotsPerEpoch)
	slotAssigIndexes := make(map[types.Slot][]types.ValidatorIndex, params.BeaconConfig().SlotsPerEpoch)
	ctx := s.ctx
	state := s.headState(ctx)
	epoch := slots.ToEpoch(slot)
	validatorIndexToCommittee, _, err := helpers.CommitteeAssignments(ctx, state, epoch)
	if err != nil {
		return []common.Address{}, err
	}

	for inx, val := range validatorIndexToCommittee {
		if slotAssigIndexes[val.AttesterSlot] == nil {
			slotAssigIndexes[val.AttesterSlot] = []types.ValidatorIndex{}
		}
		if creatorsAssig[val.AttesterSlot] == nil {
			creatorsAssig[val.AttesterSlot] = []common.Address{}
		}
		//check index in slot
		isCreator := false
		for i, vix := range val.Committee {
			if i >= int(params.BeaconConfig().MaxCreatorsPerSlot) {
				break
			}
			if inx == vix {
				isCreator = true
			}
		}
		if isCreator {
			// skip already added
			for _, vix := range slotAssigIndexes[val.AttesterSlot] {
				if inx == vix {
					isCreator = false
				}
			}
		}
		if isCreator {
			slotAssigIndexes[val.AttesterSlot] = append(slotAssigIndexes[val.AttesterSlot], inx)
			// retrieve and set creator address
			validator, err := s.headValidatorAtIndex(inx)
			if err != nil {
				log.WithError(err).Errorf("Get validator data failed: index=%v", inx)
				continue
			}
			// Withdrawal address uses as gwat coinbase
			address := common.BytesToAddress(validator.WithdrawalCredentials()[12:])
			creatorsAssig[val.AttesterSlot] = append(creatorsAssig[val.AttesterSlot], address)
		}
	}
	s.creators.assignment = creatorsAssig
	return s.creators.assignment[slot], nil
}
