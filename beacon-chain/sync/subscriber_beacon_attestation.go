package sync

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
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
		logrus.WithError(
			fmt.Errorf("message was not type *eth.Attestation, type=%T", msg),
		).WithFields(logrus.Fields{
			"curSlot": slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())), // nolint
		}).Error("Atts: incoming: handler: bad type")
		return fmt.Errorf("message was not type *eth.Attestation, type=%T", msg)
	}

	if a.Data == nil {
		logrus.WithError(errors.New("nil attestation")).WithFields(logrus.Fields{
			"curSlot": slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())), // nolint
		}).Error("Atts: incoming: handler: nil data")
		return errors.New("nil attestation")
	}
	s.setSeenCommitteeIndicesSlot(a.Data.Slot, a.Data.CommitteeIndex, a.AggregationBits)

	exists, err := s.cfg.attPool.HasAggregatedAttestation(a)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"curSlot": slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())), // nolint
		}).Error("Atts: incoming: handler: Could not determine if attestation pool has this atttestation")
		return errors.Wrap(err, "Could not determine if attestation pool has this atttestation")
	}
	if exists {
		return nil
	}

	log.WithFields(logrus.Fields{
		"curSlot":       slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())), // nolint
		"pv.AggrBits":   fmt.Sprintf("%#x", a.AggregationBits),
		"pv.Data.Slot":  a.Data.Slot,
		"pv.Data.Index": fmt.Sprintf("%#x", a.Data.BeaconBlockRoot),
	}).Debug("Atts: incoming: handler: success")

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

func (_ *Service) prevotingSubnetIndices(currentSlot types.Slot) []uint64 {
	endEpoch := slots.ToEpoch(currentSlot) + 1
	endSlot := params.BeaconConfig().SlotsPerEpoch.Mul(uint64(endEpoch))
	var commIds []uint64
	for i := currentSlot; i <= endSlot; i++ {
		commIds = append(commIds, cache.SubnetIDs.GetPrevotingSubnetIDs(i)...)
	}
	return slice.SetUint64(commIds)
}
