package p2p

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"time"

	ssz "github.com/ferranbt/fastssz"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/altair"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/hash"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/monitoring/tracing"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	"go.opencensus.io/trace"
	"google.golang.org/protobuf/proto"
)

// ErrMessageNotMapped occurs on a Broadcast attempt when a message has not been defined in the
// GossipTypeMapping.
var ErrMessageNotMapped = errors.New("message type is not mapped to a PubSub topic")

// Broadcast a message to the p2p network, the message is assumed to be
// broadcasted to the current fork.
func (s *Service) Broadcast(ctx context.Context, msg proto.Message) error {
	ctx, span := trace.StartSpan(ctx, "p2p.Broadcast")
	defer span.End()

	twoSlots := time.Duration(2*params.BeaconConfig().SecondsPerSlot) * time.Second
	ctx, cancel := context.WithTimeout(ctx, twoSlots)
	defer cancel()

	forkDigest, err := s.currentForkDigest()
	if err != nil {
		err := errors.Wrap(err, "could not retrieve fork digest")
		tracing.AnnotateError(span, err)
		return err
	}

	topic, ok := GossipTypeMapping[reflect.TypeOf(msg)]
	if !ok {
		tracing.AnnotateError(span, ErrMessageNotMapped)
		return ErrMessageNotMapped
	}
	castMsg, ok := msg.(ssz.Marshaler)
	if !ok {
		return errors.Errorf("message of %T does not support marshaller interface", msg)
	}
	return s.broadcastObject(ctx, castMsg, fmt.Sprintf(topic, forkDigest))
}

// BroadcastAttestation broadcasts an attestation to the p2p network, the message is assumed to be
// broadcasted to the current fork.
func (s *Service) BroadcastAttestation(ctx context.Context, subnet uint64, att *ethpb.Attestation) error {
	if att == nil {
		return errors.New("attempted to broadcast nil attestation")
	}
	ctx, span := trace.StartSpan(ctx, "p2p.BroadcastAttestation")
	defer span.End()
	forkDigest, err := s.currentForkDigest()
	if err != nil {
		err := errors.Wrap(err, "could not retrieve fork digest")
		tracing.AnnotateError(span, err)
		return err
	}

	// Non-blocking broadcast, with attempts to discover a subnet peer if none available.
	go s.internalBroadcastAttestation(ctx, subnet, att, forkDigest)

	return nil
}

// BroadcastPrevoting broadcasts a prevote message to the p2p network, the message is assumed to be
// broadcasted to the current fork.
func (s *Service) BroadcastPrevoting(ctx context.Context, subnet uint64, prevote *ethpb.PreVote) error {
	ctx, span := trace.StartSpan(ctx, "p2p.BroadcastPrevoting")
	defer span.End()

	log.WithFields(logrus.Fields{
		"subnet":             subnet,
		"curSlot":            slots.CurrentSlot(uint64(s.genesisTime.Unix())),
		"pv.AggregationBits": fmt.Sprintf("%b", prevote.AggregationBits),
		"pv.Data.Slot":       prevote.Data.Slot,
		"pv.Data.Index":      prevote.Data.Index,
	}).Debug("Prevote: BroadcastPrevoting")

	forkDigest, err := s.currentForkDigest()
	if err != nil {
		err := errors.Wrap(err, "could not retrieve fork digest")
		tracing.AnnotateError(span, err)
		return err
	}

	// Non-blocking broadcast, with attempts to discover a subnet peer if none available.
	go s.broadcastPrevoting(ctx, subnet, prevote, forkDigest)

	return nil
}

// BroadcastSyncCommitteeMessage broadcasts a sync committee message to the p2p network, the message is assumed to be
// broadcasted to the current fork.
func (s *Service) BroadcastSyncCommitteeMessage(ctx context.Context, subnet uint64, sMsg *ethpb.SyncCommitteeMessage) error {
	if sMsg == nil {
		return errors.New("attempted to broadcast nil sync committee message")
	}
	ctx, span := trace.StartSpan(ctx, "p2p.BroadcastSyncCommitteeMessage")
	defer span.End()
	forkDigest, err := s.currentForkDigest()
	if err != nil {
		err := errors.Wrap(err, "could not retrieve fork digest")
		tracing.AnnotateError(span, err)
		return err
	}

	// Non-blocking broadcast, with attempts to discover a subnet peer if none available.
	go s.broadcastSyncCommittee(ctx, subnet, sMsg, forkDigest)

	return nil
}

func (s *Service) internalBroadcastAttestation(ctx context.Context, subnet uint64, att *ethpb.Attestation, forkDigest [4]byte) {
	_, span := trace.StartSpan(ctx, "p2p.internalBroadcastAttestation")
	defer span.End()
	ctx = trace.NewContext(context.Background(), span) // clear parent context / deadline.

	//log.WithFields(logrus.Fields{
	//	"subnet":       subnet,
	//	"att.AggrBits": fmt.Sprintf("%#x", att.AggregationBits),
	//	"att.Slot":     att.Data.Slot,
	//	"att.root":     fmt.Sprintf("%#x", att.Data.BeaconBlockRoot),
	//	"curSlot":      slots.CurrentSlot(uint64(s.genesisTime.Unix())),
	//}).Debug("Atts: broadcastAttestation")

	oneEpoch := time.Duration(1*params.BeaconConfig().SlotsPerEpoch.Mul(params.BeaconConfig().SecondsPerSlot)) * time.Second
	ctx, cancel := context.WithTimeout(ctx, oneEpoch)
	defer cancel()

	// Ensure we have peers with this subnet.
	s.subnetLocker(subnet).RLock()
	hasPeer := s.hasPeerWithSubnet(attestationToTopic(subnet, forkDigest))
	s.subnetLocker(subnet).RUnlock()

	span.AddAttributes(
		trace.BoolAttribute("hasPeer", hasPeer),
		trace.Int64Attribute("slot", int64(att.Data.Slot)), // lint:ignore uintcast -- It's safe to do this for tracing.
		trace.Int64Attribute("subnet", int64(subnet)),      // lint:ignore uintcast -- It's safe to do this for tracing.
	)

	//log.WithFields(logrus.Fields{
	//	"subnet":        subnet,
	//	"att.AggrBits":  fmt.Sprintf("%#x", att.AggregationBits),
	//	"curSlot":       slots.CurrentSlot(uint64(s.genesisTime.Unix())),
	//	"pv.Data.Slot":  att.Data.Slot,
	//	"pv.Data.Index": att.Data.CommitteeIndex,
	//	"hasPeer":       hasPeer,
	//	//"pv.Candidates":      fmt.Sprintf("%#x", prevote.Data.Candidates),
	//}).Debug("Atts: broadcastAttestation")

	if !hasPeer {
		attestationBroadcastAttempts.Inc()
		if err := func() error {
			s.subnetLocker(subnet).Lock()
			defer s.subnetLocker(subnet).Unlock()
			ok, err := s.FindPeersWithSubnet(ctx, attestationToTopic(subnet, forkDigest), subnet, 1)

			//log.WithError(err).WithFields(logrus.Fields{
			//	"topic":        attestationToTopic(subnet, forkDigest),
			//	"ok":           ok,
			//	"subnet":       subnet,
			//	"curSlot":      slots.CurrentSlot(uint64(s.genesisTime.Unix())),
			//	"att.AggrBits": fmt.Sprintf("%#x", att.AggregationBits),
			//	"pv.Data.Slot": att.Data.Slot,
			//	//"pv.Data.Index":      prevote.Data.Index,
			//	//"hasPeer":            hasPeer,
			//	//"pv.Candidates":      fmt.Sprintf("%#x", prevote.Data.Candidates),
			//}).Debug("Atts: broadcastAttestation: FindPeersWithSubnet")

			if err != nil {
				return err
			}
			if ok {
				savedAttestationBroadcasts.Inc()
				return nil
			}
			return errors.New("failed to find peers for subnet")
		}(); err != nil {
			log.WithError(err).Error("Failed to find peers")
			tracing.AnnotateError(span, err)
		}
	}
	// In the event our attestation is outdated and beyond the
	// acceptable threshold, we exit early and do not broadcast it.
	currSlot := slots.CurrentSlot(uint64(s.genesisTime.Unix()))
	if att.Data.Slot+params.BeaconConfig().SlotsPerEpoch < currSlot {
		log.WithFields(logrus.Fields{
			"attestationSlot": att.Data.Slot,
			"currentSlot":     currSlot,
		}).Warning("Attestation is too old to broadcast, discarding it")
		return
	}

	if err := s.broadcastObject(ctx, att, attestationToTopic(subnet, forkDigest)); err != nil {
		log.WithError(err).Error("Atts: Failed to broadcast attestation")
		tracing.AnnotateError(span, err)
	}

	//log.WithFields(logrus.Fields{
	//	"topic":        attestationToTopic(subnet, forkDigest),
	//	"subnet":       subnet,
	//	"att.AggrBits": fmt.Sprintf("%#x", att.AggregationBits),
	//	"att.Slot":     att.Data.Slot,
	//	"att.root":     fmt.Sprintf("%#x", att.Data.BeaconBlockRoot),
	//	"curSlot":      slots.CurrentSlot(uint64(s.genesisTime.Unix())),
	//}).Debug("Atts: broadcastAttestation: broadcastObject 999")
}

func (s *Service) broadcastSyncCommittee(ctx context.Context, subnet uint64, sMsg *ethpb.SyncCommitteeMessage, forkDigest [4]byte) {
	_, span := trace.StartSpan(ctx, "p2p.broadcastSyncCommittee")
	defer span.End()
	ctx = trace.NewContext(context.Background(), span) // clear parent context / deadline.

	oneSlot := time.Duration(1*params.BeaconConfig().SecondsPerSlot) * time.Second
	ctx, cancel := context.WithTimeout(ctx, oneSlot)
	defer cancel()

	// Ensure we have peers with this subnet.
	// This adds in a special value to the subnet
	// to ensure that we can re-use the same subnet locker.
	wrappedSubIdx := subnet + syncLockerVal
	s.subnetLocker(wrappedSubIdx).RLock()
	hasPeer := s.hasPeerWithSubnet(syncCommitteeToTopic(subnet, forkDigest))
	s.subnetLocker(wrappedSubIdx).RUnlock()

	span.AddAttributes(
		trace.BoolAttribute("hasPeer", hasPeer),
		trace.Int64Attribute("slot", int64(sMsg.Slot)), // lint:ignore uintcast -- It's safe to do this for tracing.
		trace.Int64Attribute("subnet", int64(subnet)),  // lint:ignore uintcast -- It's safe to do this for tracing.
	)

	if !hasPeer {
		syncCommitteeBroadcastAttempts.Inc()
		if err := func() error {
			s.subnetLocker(wrappedSubIdx).Lock()
			defer s.subnetLocker(wrappedSubIdx).Unlock()
			ok, err := s.FindPeersWithSubnet(ctx, syncCommitteeToTopic(subnet, forkDigest), subnet, 1)
			if err != nil {
				return err
			}
			if ok {
				savedSyncCommitteeBroadcasts.Inc()
				return nil
			}
			return errors.New("failed to find peers for subnet")
		}(); err != nil {
			log.WithError(err).Error("Failed to find peers")
			tracing.AnnotateError(span, err)
		}
	}
	// In the event our sync message is outdated and beyond the
	// acceptable threshold, we exit early and do not broadcast it.
	if err := altair.ValidateSyncMessageTime(sMsg.Slot, s.genesisTime, params.BeaconNetworkConfig().MaximumGossipClockDisparity); err != nil {
		log.WithError(err).Warn("Sync Committee Message is too old to broadcast, discarding it")
		return
	}

	if err := s.broadcastObject(ctx, sMsg, syncCommitteeToTopic(subnet, forkDigest)); err != nil {
		log.WithError(err).Error("Failed to broadcast sync committee message")
		tracing.AnnotateError(span, err)
	}
}

func (s *Service) broadcastPrevoting(ctx context.Context, subnet uint64, prevote *ethpb.PreVote, forkDigest [4]byte) {
	ctx, span := trace.StartSpan(ctx, "p2p.broadcastPrevoting")
	defer span.End()
	ctx = trace.NewContext(context.Background(), span) // clear parent context / deadline.

	oneSlot := time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second
	ctx, cancel := context.WithTimeout(ctx, oneSlot)
	defer cancel()

	// Ensure we have peers with this subnet.
	// This adds in a special value to the subnet
	// to ensure that we can re-use the same subnet locker.
	s.subnetLocker(subnet).RLock()
	ptt := prevoteToTopic(subnet, forkDigest)
	hasPeer := s.hasPeerWithSubnet(ptt)
	s.subnetLocker(subnet).RUnlock()

	span.AddAttributes(
		trace.BoolAttribute("hasPeer", hasPeer),
		trace.Int64Attribute("slot", int64(prevote.Data.Slot)), // lint:ignore uintcast -- It's safe to do this for tracing.
		trace.Int64Attribute("subnet", int64(subnet)),          // lint:ignore uintcast -- It's safe to do this for tracing.
	)

	log.WithFields(logrus.Fields{
		"0:subnet":           subnet,
		"1:hasPeer":          hasPeer,
		"curSlot":            slots.CurrentSlot(uint64(s.genesisTime.Unix())),
		"pv.AggregationBits": fmt.Sprintf("%b", prevote.AggregationBits),
		"pv.Data.Slot":       prevote.Data.Slot,
		"pv.Data.Index":      prevote.Data.Index,
	}).Debug("Prevote: BroadcastPrevoting: broadcastPrevoting 1")

	if !hasPeer {
		prevoteBroadcastAttempts.Inc()
		if err := func() error {
			s.subnetLocker(subnet).Lock()
			defer s.subnetLocker(subnet).Unlock()
			ok, err := s.FindPeersWithSubnet(ctx, ptt, subnet, 1)
			if err != nil {
				return err
			}
			if ok {
				savedPrevoteBroadcasts.Inc()
				return nil
			}
			return errors.New("failed to find peers for subnet")
		}(); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"subnet":             subnet,
				"hasPeer":            hasPeer,
				"curSlot":            slots.CurrentSlot(uint64(s.genesisTime.Unix())),
				"pv.AggregationBits": fmt.Sprintf("%b", prevote.AggregationBits),
				"pv.Data.Slot":       prevote.Data.Slot,
				"pv.Data.Index":      prevote.Data.Index,
			}).Error("Prevote: BroadcastPrevoting: broadcastPrevoting failed")
			tracing.AnnotateError(span, err)
		}
	}
	// In the event our prevote is outdated and beyond the
	// acceptable threshold, we exit early and do not broadcast it.
	currSlot := slots.CurrentSlot(uint64(s.genesisTime.Unix()))

	if prevote.Data.Slot <= currSlot {
		log.Warnf("Prevote is too old to broadcast, discarding it. Current Slot: %d , Prevote Slot: %d", currSlot, prevote.Data.Slot)
		return
	}

	log.WithFields(logrus.Fields{
		"subnet":             subnet,
		"curSlot":            slots.CurrentSlot(uint64(s.genesisTime.Unix())),
		"pv.AggregationBits": fmt.Sprintf("%#x", prevote.AggregationBits),
		"pv.Data.Slot":       prevote.Data.Slot,
		"pv.Data.Index":      prevote.Data.Index,
	}).Debug("Prevote: BroadcastPrevoting: broadcastPrevoting 2")

	if err := s.broadcastObject(ctx, prevote, ptt); err != nil {
		log.WithError(err).Error("Failed to broadcast prevote")
		tracing.AnnotateError(span, err)
	}
}

// method to broadcast messages to other peers in our gossip mesh.
func (s *Service) broadcastObject(ctx context.Context, obj ssz.Marshaler, topic string) error {
	ctx, span := trace.StartSpan(ctx, "p2p.broadcastObject")
	defer span.End()

	span.AddAttributes(trace.StringAttribute("topic", topic))

	buf := new(bytes.Buffer)
	if _, err := s.Encoding().EncodeGossip(buf, obj); err != nil {
		err := errors.Wrap(err, "could not encode message")
		tracing.AnnotateError(span, err)
		return err
	}

	if span.IsRecordingEvents() {
		id := hash.FastSum64(buf.Bytes())
		messageLen := int64(buf.Len())
		// lint:ignore uintcast -- It's safe to do this for tracing.
		iid := int64(id)
		span.AddMessageSendEvent(iid, messageLen /*uncompressed*/, messageLen /*compressed*/)
	}
	if err := s.PublishToTopic(ctx, topic+s.Encoding().ProtocolSuffix(), buf.Bytes()); err != nil {
		err := errors.Wrap(err, "could not publish message")
		tracing.AnnotateError(span, err)
		return err
	}
	return nil
}

func attestationToTopic(subnet uint64, forkDigest [4]byte) string {
	return fmt.Sprintf(AttestationSubnetTopicFormat, forkDigest, subnet)
}

func syncCommitteeToTopic(subnet uint64, forkDigest [4]byte) string {
	return fmt.Sprintf(SyncCommitteeSubnetTopicFormat, forkDigest, subnet)
}

func prevoteToTopic(subnet uint64, forkDigest [4]byte) string {
	return fmt.Sprintf(PrevoteSubnetTopicFormat, forkDigest, subnet)
}
