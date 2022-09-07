package blockchain

import (
	"sort"
	"sync"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/config/params"
	"github.com/prysmaticlabs/prysm/time/slots"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

// This defines the current chain service's view of creators.
type creatorsAssignment struct {
	assignment map[types.Slot][]gwatCommon.Address // creators' assignment by slot.
	lock       sync.RWMutex
}

type listValIx []types.ValidatorIndex

func (v listValIx) Len() int           { return len(v) }
func (v listValIx) Less(i, j int) bool { return v[i] < v[j] }
func (v listValIx) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

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
	creatorsAssig := make(map[types.Slot][]gwatCommon.Address, params.BeaconConfig().SlotsPerEpoch)
	slotAssigIndexes := make(map[types.Slot][]types.ValidatorIndex, params.BeaconConfig().SlotsPerEpoch)
	ctx := s.ctx
	state := s.headState(ctx)
	epoch := slots.ToEpoch(slot)
	validatorIndexToCommittee, _, err := helpers.CommitteeAssignments(ctx, state, epoch)
	if err != nil {
		return []gwatCommon.Address{}, err
	}

	vIxs := listValIx{}
	for inx, _ := range validatorIndexToCommittee {
		vIxs = append(vIxs, inx)
	}
	sort.Sort(vIxs)
	for _, inx := range vIxs {
		val := validatorIndexToCommittee[inx]

		if slotAssigIndexes[val.AttesterSlot] == nil {
			slotAssigIndexes[val.AttesterSlot] = []types.ValidatorIndex{}
		}
		if creatorsAssig[val.AttesterSlot] == nil {
			creatorsAssig[val.AttesterSlot] = []gwatCommon.Address{}
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
			slotAssigIndexes[val.AttesterSlot] = append(slotAssigIndexes[val.AttesterSlot], inx)
			// retrieve and set creator address
			validator, err := s.headValidatorAtIndex(inx)
			if err != nil {
				log.WithError(err).Errorf("Get validator data failed: index=%v", inx)
				continue
			}
			// Withdrawal address uses as gwat coinbase
			address := gwatCommon.BytesToAddress(validator.WithdrawalCredentials()[12:])
			// skip already added
			for _, addr := range creatorsAssig[val.AttesterSlot] {
				if address == addr {
					isCreator = false
					break
				}
			}
			if isCreator {
				creatorsAssig[val.AttesterSlot] = gwatCommon.SortAddresses(append(creatorsAssig[val.AttesterSlot], address))
			}
		}
	}
	s.creators.assignment = creatorsAssig
	return s.creators.assignment[slot], nil
}
