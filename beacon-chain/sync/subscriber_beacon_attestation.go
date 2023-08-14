package sync

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/container/slice"
	eth "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	"google.golang.org/protobuf/proto"
)

func (s *Service) committeeIndexBeaconAttestationSubscriber(_ context.Context, msg proto.Message) error {
	a, ok := msg.(*eth.Attestation)
	if !ok {
		return fmt.Errorf("message was not type *eth.Attestation, type=%T", msg)
	}

	if a.Data == nil {
		return errors.New("nil attestation")
	}
	s.setSeenCommitteeIndicesSlot(a.Data.Slot, a.Data.CommitteeIndex, a.AggregationBits)

	exists, err := s.cfg.attPool.HasAggregatedAttestation(a)
	if err != nil {
		return errors.Wrap(err, "Could not determine if attestation pool has this atttestation")
	}
	if exists {
		return nil
	}

	return s.cfg.attPool.SaveUnaggregatedAttestation(a)
}

func (_ *Service) persistentSubnetIndices() []uint64 {
	return cache.SubnetIDs.GetAllSubnets()
}

func (_ *Service) aggregatorSubnetIndices(currentSlot types.Slot) []uint64 {
	endEpoch := slots.ToEpoch(currentSlot) + 1
	endSlot := params.BeaconConfig().SlotsPerEpoch.Mul(uint64(endEpoch))
	var commIds []uint64
	for i := currentSlot; i <= endSlot; i++ {
		commIds = append(commIds, cache.SubnetIDs.GetAggregatorSubnetIDs(i)...)
	}
	return slice.SetUint64(commIds)
}

func (_ *Service) attesterSubnetIndices(currentSlot types.Slot) []uint64 {
	endEpoch := slots.ToEpoch(currentSlot) + 1
	endSlot := params.BeaconConfig().SlotsPerEpoch.Mul(uint64(endEpoch))
	var commIds []uint64
	for i := currentSlot; i <= endSlot; i++ {
		commIds = append(commIds, cache.SubnetIDs.GetAttesterSubnetIDs(i)...)
	}
	return slice.SetUint64(commIds)
}

func (_ *Service) proposerSubnetIndices(currentSlot types.Slot) []uint64 {
	endEpoch := slots.ToEpoch(currentSlot) + 1
	endSlot := params.BeaconConfig().SlotsPerEpoch.Mul(uint64(endEpoch))
	var commIds []uint64
	for i := currentSlot; i <= endSlot; i++ {
		commIds = append(commIds, cache.SubnetIDs.GetProposerSubnetIDs(i)...)
	}
	return slice.SetUint64(commIds)
}
