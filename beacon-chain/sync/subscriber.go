package sync

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p/peers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd/beacon-chain/flags"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/container/slice"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/monitoring/tracing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/network/forks"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/messagehandler"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	"go.opencensus.io/trace"
	"google.golang.org/protobuf/proto"
)

const pubsubMessageTimeout = 30 * time.Second

// wrappedVal represents a gossip validator which also returns an error along with the result.
type wrappedVal func(context.Context, peer.ID, *pubsub.Message) (pubsub.ValidationResult, error)

// subHandler represents handler for a given subscription.
type subHandler func(context.Context, proto.Message) error

// noopValidator is a no-op that only decodes the message, but does not check its contents.
func (s *Service) noopValidator(_ context.Context, _ peer.ID, msg *pubsub.Message) (pubsub.ValidationResult, error) {
	m, err := s.decodePubsubMessage(msg)
	if err != nil {
		log.WithError(err).Debug("Could not decode message")
		return pubsub.ValidationReject, nil
	}
	msg.ValidatorData = m
	return pubsub.ValidationAccept, nil
}

// Register PubSub subscribers
func (s *Service) registerSubscribers(epoch types.Epoch, digest [4]byte) {
	s.subscribe(
		p2p.BlockSubnetTopicFormat,
		s.validateBeaconBlockPubSub,
		s.beaconBlockSubscriber,
		digest,
	)
	s.subscribe(
		p2p.AggregateAndProofSubnetTopicFormat,
		s.validateAggregateAndProof,
		s.beaconAggregateProofSubscriber,
		digest,
	)
	s.subscribe(
		p2p.ExitSubnetTopicFormat,
		s.validateVoluntaryExit,
		s.voluntaryExitSubscriber,
		digest,
	)
	s.subscribe(
		p2p.ProposerSlashingSubnetTopicFormat,
		s.validateProposerSlashing,
		s.proposerSlashingSubscriber,
		digest,
	)
	s.subscribe(
		p2p.AttesterSlashingSubnetTopicFormat,
		s.validateAttesterSlashing,
		s.attesterSlashingSubscriber,
		digest,
	)
	if flags.Get().SubscribeToAllSubnets {
		s.subscribeStaticWithSubnets(
			p2p.AttestationSubnetTopicFormat,
			s.validateCommitteeIndexBeaconAttestation,   /* validator */
			s.committeeIndexBeaconAttestationSubscriber, /* message handler */
			digest,
		)
	} else {
		s.subscribeDynamicWithSubnets(
			p2p.AttestationSubnetTopicFormat,
			s.validateCommitteeIndexBeaconAttestation,   /* validator */
			s.committeeIndexBeaconAttestationSubscriber, /* message handler */
			digest,
		)
	}
	// Altair Fork Version
	if epoch >= params.BeaconConfig().AltairForkEpoch {
		s.subscribe(
			p2p.SyncContributionAndProofSubnetTopicFormat,
			s.validateSyncContributionAndProof,
			s.syncContributionAndProofSubscriber,
			digest,
		)
		if flags.Get().SubscribeToAllSubnets {
			s.subscribeStaticWithSyncSubnets(
				p2p.SyncCommitteeSubnetTopicFormat,
				s.validateSyncCommitteeMessage,   /* validator */
				s.syncCommitteeMessageSubscriber, /* message handler */
				digest,
			)
		} else {
			s.subscribeDynamicWithSyncSubnets(
				p2p.SyncCommitteeSubnetTopicFormat,
				s.validateSyncCommitteeMessage,   /* validator */
				s.syncCommitteeMessageSubscriber, /* message handler */
				digest,
			)
		}
	}
}

// subscribe to a given topic with a given validator and subscription handler.
// The base protobuf message is used to initialize new messages for decoding.
func (s *Service) subscribe(topic string, validator wrappedVal, handle subHandler, digest [4]byte) *pubsub.Subscription {
	genRoot := s.cfg.chain.GenesisValidatorsRoot()
	_, e, err := forks.RetrieveForkDataFromDigest(digest, genRoot[:])
	if err != nil {
		// Impossible condition as it would mean digest does not exist.
		panic(err)
	}
	base := p2p.GossipTopicMappings(topic, e)
	if base == nil {
		// Impossible condition as it would mean topic does not exist.
		panic(fmt.Sprintf("%s is not mapped to any message in GossipTopicMappings", topic))
	}
	return s.subscribeWithBase(s.addDigestToTopic(topic, digest), validator, handle)
}

func (s *Service) subscribeWithBase(topic string, validator wrappedVal, handle subHandler) *pubsub.Subscription {
	topic += s.cfg.p2p.Encoding().ProtocolSuffix()
	log := log.WithField("topic", topic)

	// Do not resubscribe already seen subscriptions.
	ok := s.subHandler.topicExists(topic)
	if ok {
		log.Debugf("Provided topic already has an active subscription running: %s", topic)
		return nil
	}

	if err := s.cfg.p2p.PubSub().RegisterTopicValidator(s.wrapAndReportValidation(topic, validator)); err != nil {
		log.WithError(err).Error("Could not register validator for topic")
		return nil
	}

	sub, err := s.cfg.p2p.SubscribeToTopic(topic)
	if err != nil {
		// Any error subscribing to a PubSub topic would be the result of a misconfiguration of
		// libp2p PubSub library or a subscription request to a topic that fails to match the topic
		// subscription filter.
		log.WithError(err).Error("Could not subscribe topic")
		return nil
	}
	s.subHandler.addTopic(sub.Topic(), sub)

	// Pipeline decodes the incoming subscription data, runs the validation, and handles the
	// message.
	pipeline := func(msg *pubsub.Message) {
		ctx, cancel := context.WithTimeout(s.ctx, pubsubMessageTimeout)
		defer cancel()
		ctx, span := trace.StartSpan(ctx, "sync.pubsub")
		defer span.End()

		defer func() {
			if r := recover(); r != nil {
				tracing.AnnotateError(span, fmt.Errorf("panic occurred: %v", r))
				log.WithField("error", r).Error("Panic occurred")
				debug.PrintStack()
			}
		}()

		span.AddAttributes(trace.StringAttribute("topic", topic))

		if msg.ValidatorData == nil {
			log.Error("Received nil message on pubsub")
			messageFailedProcessingCounter.WithLabelValues(topic).Inc()
			return
		}

		if err := handle(ctx, msg.ValidatorData.(proto.Message)); err != nil {
			tracing.AnnotateError(span, err)
			log.WithError(err).Error("Could not handle p2p pubsub")
			messageFailedProcessingCounter.WithLabelValues(topic).Inc()
			return
		}
	}

	// The main message loop for receiving incoming messages from this subscription.
	messageLoop := func() {
		for {
			msg, err := sub.Next(s.ctx)
			if err != nil {
				// This should only happen when the context is cancelled or subscription is cancelled.
				if err != pubsub.ErrSubscriptionCancelled { // Only log a warning on unexpected errors.
					log.WithError(err).Warn("Subscription next failed")
				}
				// Cancel subscription in the event of an error, as we are
				// now exiting topic event loop.
				sub.Cancel()
				return
			}

			if msg.ReceivedFrom == s.cfg.p2p.PeerID() {
				continue
			}

			go pipeline(msg)
		}
	}

	go messageLoop()
	log.WithField("topic", topic).Info("Subscribed to topic")
	return sub
}

// Wrap the pubsub validator with a metric monitoring function. This function increments the
// appropriate counter if the particular message fails to validate.
func (s *Service) wrapAndReportValidation(topic string, v wrappedVal) (string, pubsub.ValidatorEx) {
	return topic, func(ctx context.Context, pid peer.ID, msg *pubsub.Message) (res pubsub.ValidationResult) {
		defer messagehandler.HandlePanic(ctx, msg)
		res = pubsub.ValidationIgnore // Default: ignore any message that panics.
		ctx, cancel := context.WithTimeout(ctx, pubsubMessageTimeout)
		defer cancel()
		messageReceivedCounter.WithLabelValues(topic).Inc()
		if msg.Topic == nil {
			messageFailedValidationCounter.WithLabelValues(topic).Inc()
			return pubsub.ValidationReject
		}
		// Ignore any messages received before chainstart.
		if s.chainStarted.IsNotSet() {
			messageIgnoredValidationCounter.WithLabelValues(topic).Inc()
			return pubsub.ValidationIgnore
		}
		retDigest, err := p2p.ExtractGossipDigest(topic)
		if err != nil {
			log.WithField("topic", topic).Errorf("Invalid topic format of pubsub topic: %v", err)
			return pubsub.ValidationIgnore
		}
		currDigest, err := s.currentForkDigest()
		if err != nil {
			log.WithField("topic", topic).Errorf("Unable to retrieve fork data: %v", err)
			return pubsub.ValidationIgnore
		}
		if currDigest != retDigest {
			log.WithField("topic", topic).Debugf("Received message from outdated fork digest %#x", retDigest)
			return pubsub.ValidationIgnore
		}
		b, err := v(ctx, pid, msg)
		if b == pubsub.ValidationReject {
			log.WithError(err).WithFields(logrus.Fields{
				"topic":        topic,
				"multiaddress": multiAddr(pid, s.cfg.p2p.Peers()),
				"peer id":      pid.String(),
				"agent":        agentString(pid, s.cfg.p2p.Host()),
				"gossip score": s.cfg.p2p.Peers().Scorers().GossipScorer().Score(pid),
			}).Debugf("Gossip message was rejected")
			messageFailedValidationCounter.WithLabelValues(topic).Inc()
		}
		if b == pubsub.ValidationIgnore {
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"topic":        topic,
					"multiaddress": multiAddr(pid, s.cfg.p2p.Peers()),
					"peer id":      pid.String(),
					"agent":        agentString(pid, s.cfg.p2p.Host()),
					"gossip score": s.cfg.p2p.Peers().Scorers().GossipScorer().Score(pid),
				}).Debugf("Gossip message was ignored")
			}
			messageIgnoredValidationCounter.WithLabelValues(topic).Inc()
		}
		return b
	}
}

// subscribe to a static subnet  with the given topic and index.A given validator and subscription handler is
// used to handle messages from the subnet. The base protobuf message is used to initialize new messages for decoding.
func (s *Service) subscribeStaticWithSubnets(topic string, validator wrappedVal, handle subHandler, digest [4]byte) {
	genRoot := s.cfg.chain.GenesisValidatorsRoot()
	_, e, err := forks.RetrieveForkDataFromDigest(digest, genRoot[:])
	if err != nil {
		// Impossible condition as it would mean digest does not exist.
		panic(err)
	}
	base := p2p.GossipTopicMappings(topic, e)
	if base == nil {
		// Impossible condition as it would mean topic does not exist.
		panic(fmt.Sprintf("%s is not mapped to any message in GossipTopicMappings", topic))
	}
	for i := uint64(0); i < params.BeaconNetworkConfig().AttestationSubnetCount; i++ {
		s.subscribeWithBase(s.addDigestAndIndexToTopic(topic, digest, i), validator, handle)
	}
	genesis := s.cfg.chain.GenesisTime()
	ticker := slots.NewSlotTicker(genesis, params.BeaconConfig().SecondsPerSlot)

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				ticker.Done()
				return
			case <-ticker.C():
				if s.chainStarted.IsSet() && s.cfg.initialSync.Syncing() {

					//todo RM
					log.WithError(err).WithFields(logrus.Fields{
						"digest":    fmt.Sprintf("%#x", digest),
						"topic":     topic,
						"s.curSlot": s.cfg.chain.CurrentSlot(),
						"curSlot":   slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
					}).Info("Prevote: subscribeStaticWithSubnets: continue 000 ")

					continue
				}
				valid, err := isDigestValid(digest, genesis, genRoot)
				if err != nil {
					log.Error(err)

					//todo RM
					log.WithError(err).WithFields(logrus.Fields{
						"digest":        fmt.Sprintf("%#x", digest),
						"isDigestValid": valid,
						"topic":         topic,
						"s.curSlot":     s.cfg.chain.CurrentSlot(),
						"curSlot":       slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
					}).Info("Prevote: subscribeStaticWithSubnets: continue 111 ")

					continue
				}
				if !valid {

					//todo RM
					log.WithError(err).WithFields(logrus.Fields{
						"digest":        fmt.Sprintf("%#x", digest),
						"isDigestValid": valid,
						"topic":         topic,
						"s.curSlot":     s.cfg.chain.CurrentSlot(),
						"curSlot":       slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
					}).Info("Prevote: subscribeStaticWithSubnets: return 222 -> unSubscribeFromTopic ")

					log.Warnf("Attestation subnets with digest %#x are no longer valid, unsubscribing from all of them.", digest)
					// Unsubscribes from all our current subnets.
					for i := uint64(0); i < params.BeaconNetworkConfig().AttestationSubnetCount; i++ {
						fullTopic := fmt.Sprintf(topic, digest, i) + s.cfg.p2p.Encoding().ProtocolSuffix()
						s.unSubscribeFromTopic(fullTopic)
					}
					ticker.Done()
					return
				}

				//todo RM
				log.WithError(err).WithFields(logrus.Fields{
					"digest":                 fmt.Sprintf("%#x", digest),
					"isDigestValid":          valid,
					"topic":                  topic,
					"s.curSlot":              s.cfg.chain.CurrentSlot(),
					"curSlot":                slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
					"AttestationSubnetCount": params.BeaconNetworkConfig().AttestationSubnetCount,
				}).Info("Prevote: subscribeStaticWithSubnets: 333 ")

				// Check every slot that there are enough peers
				for i := uint64(0); i < params.BeaconNetworkConfig().AttestationSubnetCount; i++ {

					log.WithError(err).WithFields(logrus.Fields{
						"i":     i,
						"topic": topic,
					}).Info("Prevote: subscribeStaticWithSubnets: 444 validPeersExist")

					if !s.validPeersExist(s.addDigestAndIndexToTopic(topic, digest, i)) {
						log.Infof("No peers found subscribed to attestation gossip subnet with "+
							"committee index %d. Searching network for peers subscribed to the subnet.", i)

						pwsn, err := s.cfg.p2p.FindPeersWithSubnet(
							s.ctx,
							s.addDigestAndIndexToTopic(topic, digest, i),
							i,
							flags.Get().MinimumPeersPerSubnet,
						)

						log.WithError(err).WithFields(logrus.Fields{
							"i":     i,
							"find":  pwsn,
							"topic": topic,
						}).Info("Prevote: subscribeStaticWithSubnets: 555 FindPeersWithSubnet")

						if err != nil {
							log.WithError(err).Info("Could not search for peers")
							return
						}
					}
				}
			}
		}
	}()
}

// subscribe to a dynamically changing list of subnets. This method expects a fmt compatible
// string for the topic name and the list of subnets for subscribed topics that should be
// maintained.
func (s *Service) subscribeDynamicWithSubnets(
	topicFormat string,
	validate wrappedVal,
	handle subHandler,
	digest [4]byte,
) {
	genRoot := s.cfg.chain.GenesisValidatorsRoot()
	_, e, err := forks.RetrieveForkDataFromDigest(digest, genRoot[:])
	if err != nil {
		// Impossible condition as it would mean digest does not exist.
		panic(err)
	}
	base := p2p.GossipTopicMappings(topicFormat, e)
	if base == nil {
		panic(fmt.Sprintf("%s is not mapped to any message in GossipTopicMappings", topicFormat))
	}
	subscriptions := make(map[uint64]*pubsub.Subscription, params.BeaconConfig().MaxCommitteesPerSlot)
	genesis := s.cfg.chain.GenesisTime()
	ticker := slots.NewSlotTicker(genesis, params.BeaconConfig().SecondsPerSlot)

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				ticker.Done()

				//todo RM
				log.WithError(err).WithFields(logrus.Fields{
					"topic":     topicFormat,
					"digest":    fmt.Sprintf("%#x", digest),
					"s.curSlot": s.cfg.chain.CurrentSlot(),
					"calcSlot":  slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
				}).Info("Prevote: subscribeDynamicWithSubnets: ticker.Done() >>>>")

				return
			case currentSlot := <-ticker.C():

				//todo RM
				log.WithError(err).WithFields(logrus.Fields{
					"topic":       topicFormat,
					"digest":      fmt.Sprintf("%#x", digest),
					"currentSlot": currentSlot,
					"s.curSlot":   s.cfg.chain.CurrentSlot(),
					"calcSlot":    slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
				}).Info("Prevote: subscribeDynamicWithSubnets: currentSlot := <-ticker.C() >>>>")

				if s.chainStarted.IsSet() && s.cfg.initialSync.Syncing() {

					//todo RM
					log.WithError(err).WithFields(logrus.Fields{
						"topic":     topicFormat,
						"digest":    fmt.Sprintf("%#x", digest),
						"s.curSlot": s.cfg.chain.CurrentSlot(),
						"curSlot":   slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
					}).Info("Prevote: subscribeDynamicWithSubnets: continue 000 ")

					continue
				}
				valid, err := isDigestValid(digest, genesis, genRoot)

				//todo RM
				log.WithError(err).WithFields(logrus.Fields{
					"digest":        fmt.Sprintf("%#x", digest),
					"isDigestValid": valid,
					"topic":         topicFormat,
					"s.curSlot":     s.cfg.chain.CurrentSlot(),
					"curSlot":       slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
				}).Info("Prevote: subscribeDynamicWithSubnets: continue 111 ")

				if err != nil {
					log.Error(err)
					continue
				}
				if !valid {

					//todo RM
					log.WithError(err).WithFields(logrus.Fields{
						"digest":        fmt.Sprintf("%#x", digest),
						"isDigestValid": valid,
						"wantedSubs":    "[EMPTY]",
						"topic":         topicFormat,
						"s.curSlot":     s.cfg.chain.CurrentSlot(),
						"curSlot":       slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
					}).Info("Prevote: subscribeDynamicWithSubnets: return 222 -> reValidateSubscriptions ")

					log.Warnf("Attestation subnets with digest %#x are no longer valid, unsubscribing from all of them.", digest)
					// Unsubscribes from all our current subnets.
					s.reValidateSubscriptions(subscriptions, []uint64{}, topicFormat, digest)
					ticker.Done()
					return
				}
				wantedSubs := s.retrievePersistentSubs(currentSlot)
				// Resize as appropriate.
				s.reValidateSubscriptions(subscriptions, wantedSubs, topicFormat, digest)

				//todo RM
				log.WithError(err).WithFields(logrus.Fields{
					"digest":     fmt.Sprintf("%#x", digest),
					"wantedSubs": wantedSubs,
					"topic":      topicFormat,
					"s.curSlot":  s.cfg.chain.CurrentSlot(),
					"curSlot":    slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
				}).Info("Prevote: subscribeDynamicWithSubnets: return 333 -> subscribeAggregatorSubnet ")

				// subscribe desired aggregator subnets.
				for _, idx := range wantedSubs {
					s.subscribeAggregatorSubnet(subscriptions, idx, digest, validate, handle)
				}

				//todo RM
				log.WithError(err).WithFields(logrus.Fields{
					"digest":     fmt.Sprintf("%#x", digest),
					"wantedSubs": wantedSubs,
					"topic":      topicFormat,
					"s.curSlot":  s.cfg.chain.CurrentSlot(),
					"curSlot":    slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
				}).Info("Prevote: subscribeDynamicWithSubnets: return 444 -> subscribeAggregatorSubnet ")

				// find desired subs for attesters
				attesterSubs := s.attesterSubnetIndices(currentSlot)

				//todo RM
				log.WithError(err).WithFields(logrus.Fields{
					"digest":       fmt.Sprintf("%#x", digest),
					"attesterSubs": attesterSubs,
					"topic":        topicFormat,
					"s.curSlot":    s.cfg.chain.CurrentSlot(),
					"curSlot":      slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
				}).Info("Prevote: subscribeDynamicWithSubnets: return 555 -> lookupAttesterSubnets ")

				for _, idx := range attesterSubs {
					s.lookupAttesterSubnets(digest, idx)
				}

				//todo RM
				log.WithError(err).WithFields(logrus.Fields{
					"digest":       fmt.Sprintf("%#x", digest),
					"attesterSubs": attesterSubs,
					"topic":        topicFormat,
					"s.curSlot":    s.cfg.chain.CurrentSlot(),
					"curSlot":      slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
				}).Info("Prevote: subscribeDynamicWithSubnets: return 666 -> lookupAttesterSubnets ")

			}
		}
	}()
}

// revalidate that our currently connected subnets are valid.
func (s *Service) reValidateSubscriptions(subscriptions map[uint64]*pubsub.Subscription,
	wantedSubs []uint64, topicFormat string, digest [4]byte) {
	for k, v := range subscriptions {
		var wanted bool
		for _, idx := range wantedSubs {
			if k == idx {
				wanted = true
				break
			}
		}
		if !wanted && v != nil {
			v.Cancel()
			fullTopic := fmt.Sprintf(topicFormat, digest, k) + s.cfg.p2p.Encoding().ProtocolSuffix()
			s.unSubscribeFromTopic(fullTopic)
			delete(subscriptions, k)
		}
	}
}

// subscribe missing subnets for our aggregators.
func (s *Service) subscribeAggregatorSubnet(
	subscriptions map[uint64]*pubsub.Subscription,
	idx uint64,
	digest [4]byte,
	validate wrappedVal,
	handle subHandler,
) {

	//todo RM
	log.WithFields(logrus.Fields{
		"digest":  fmt.Sprintf("%#x", digest),
		"idx":     idx,
		"curSlot": slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
	}).Info("Prevote: subscribeAggregatorSubnet: 000")

	// do not subscribe if we have no peers in the same
	// subnet
	topic := p2p.GossipTypeMapping[reflect.TypeOf(&ethpb.Attestation{})]
	subnetTopic := fmt.Sprintf(topic, digest, idx)

	log.WithFields(logrus.Fields{
		"digest":  fmt.Sprintf("%#x", digest),
		"idx":     idx,
		"curSlot": slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
	}).Info("Prevote: subscribeAggregatorSubnet: 111")

	// check if subscription exists and if not subscribe the relevant subnet.
	_, exists := subscriptions[idx]
	if !exists {
		subscriptions[idx] = s.subscribeWithBase(subnetTopic, validate, handle)
	}

	log.WithFields(logrus.Fields{
		"digest":  fmt.Sprintf("%#x", digest),
		"idx":     idx,
		"curSlot": slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
		"exists":  exists,
	}).Info("Prevote: subscribeAggregatorSubnet: 222")

	log.WithFields(logrus.Fields{
		"digest":          fmt.Sprintf("%#x", digest),
		"idx":             idx,
		"curSlot":         slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
		"validPeersExist": s.validPeersExist(subnetTopic),
	}).Info("Prevote: subscribeAggregatorSubnet: 333")

	if !s.validPeersExist(subnetTopic) {
		log.Infof("No peers found subscribed to attestation gossip subnet with "+
			"committee index %d. Searching network for peers subscribed to the subnet.", idx)
		_, err := s.cfg.p2p.FindPeersWithSubnet(s.ctx, subnetTopic, idx, flags.Get().MinimumPeersPerSubnet)
		if err != nil {
			log.WithError(err).Error("Could not search for peers")
		}
	}

	log.WithFields(logrus.Fields{
		"digest":  fmt.Sprintf("%#x", digest),
		"idx":     idx,
		"curSlot": slots.CurrentSlot(uint64(s.cfg.chain.GenesisTime().Unix())),
	}).Info("Prevote: subscribeAggregatorSubnet: 444")

}

// subscribe missing subnets for our sync committee members.
func (s *Service) subscribeSyncSubnet(
	subscriptions map[uint64]*pubsub.Subscription,
	idx uint64,
	digest [4]byte,
	validate wrappedVal,
	handle subHandler,
) {
	// do not subscribe if we have no peers in the same
	// subnet
	topic := p2p.GossipTypeMapping[reflect.TypeOf(&ethpb.SyncCommitteeMessage{})]
	subnetTopic := fmt.Sprintf(topic, digest, idx)
	// check if subscription exists and if not subscribe the relevant subnet.
	if _, exists := subscriptions[idx]; !exists {
		subscriptions[idx] = s.subscribeWithBase(subnetTopic, validate, handle)
	}
	if !s.validPeersExist(subnetTopic) {
		log.Infof("No peers found subscribed to sync gossip subnet with "+
			"committee index %d. Searching network for peers subscribed to the subnet.", idx)
		_, err := s.cfg.p2p.FindPeersWithSubnet(s.ctx, subnetTopic, idx, flags.Get().MinimumPeersPerSubnet)
		if err != nil {
			log.WithError(err).Debug("Could not search for peers")
		}
	}
}

// subscribe to a static subnet with the given topic and index. A given validator and subscription handler is
// used to handle messages from the subnet. The base protobuf message is used to initialize new messages for decoding.
func (s *Service) subscribeStaticWithSyncSubnets(topic string, validator wrappedVal, handle subHandler, digest [4]byte) {
	genRoot := s.cfg.chain.GenesisValidatorsRoot()
	_, e, err := forks.RetrieveForkDataFromDigest(digest, genRoot[:])
	if err != nil {
		panic(err)
	}
	base := p2p.GossipTopicMappings(topic, e)
	if base == nil {
		panic(fmt.Sprintf("%s is not mapped to any message in GossipTopicMappings", topic))
	}
	for i := uint64(0); i < params.BeaconConfig().SyncCommitteeSubnetCount; i++ {
		s.subscribeWithBase(s.addDigestAndIndexToTopic(topic, digest, i), validator, handle)
	}
	genesis := s.cfg.chain.GenesisTime()
	ticker := slots.NewSlotTicker(genesis, params.BeaconConfig().SecondsPerSlot)

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				ticker.Done()
				return
			case <-ticker.C():
				if s.chainStarted.IsSet() && s.cfg.initialSync.Syncing() {
					continue
				}
				valid, err := isDigestValid(digest, genesis, genRoot)
				if err != nil {
					log.Error(err)
					continue
				}
				if !valid {
					log.Warnf("Sync subnets with digest %#x are no longer valid, unsubscribing from all of them.", digest)
					// Unsubscribes from all our current subnets.
					for i := uint64(0); i < params.BeaconConfig().SyncCommitteeSubnetCount; i++ {
						fullTopic := fmt.Sprintf(topic, digest, i) + s.cfg.p2p.Encoding().ProtocolSuffix()
						s.unSubscribeFromTopic(fullTopic)
					}
					ticker.Done()
					return
				}
				// Check every slot that there are enough peers
				for i := uint64(0); i < params.BeaconConfig().SyncCommitteeSubnetCount; i++ {
					if !s.validPeersExist(s.addDigestAndIndexToTopic(topic, digest, i)) {
						log.Infof("No peers found subscribed to sync gossip subnet with "+
							"committee index %d. Searching network for peers subscribed to the subnet.", i)
						_, err := s.cfg.p2p.FindPeersWithSubnet(
							s.ctx,
							s.addDigestAndIndexToTopic(topic, digest, i),
							i,
							flags.Get().MinimumPeersPerSubnet,
						)
						if err != nil {
							log.WithError(err).Debug("Could not search for peers")
							return
						}
					}
				}
			}
		}
	}()
}

// subscribe to a dynamically changing list of subnets. This method expects a fmt compatible
// string for the topic name and the list of subnets for subscribed topics that should be
// maintained.
func (s *Service) subscribeDynamicWithSyncSubnets(
	topicFormat string,
	validate wrappedVal,
	handle subHandler,
	digest [4]byte,
) {
	genRoot := s.cfg.chain.GenesisValidatorsRoot()
	_, e, err := forks.RetrieveForkDataFromDigest(digest, genRoot[:])
	if err != nil {
		panic(err)
	}
	base := p2p.GossipTopicMappings(topicFormat, e)
	if base == nil {
		panic(fmt.Sprintf("%s is not mapped to any message in GossipTopicMappings", topicFormat))
	}
	subscriptions := make(map[uint64]*pubsub.Subscription, params.BeaconConfig().SyncCommitteeSubnetCount)
	genesis := s.cfg.chain.GenesisTime()
	ticker := slots.NewSlotTicker(genesis, params.BeaconConfig().SecondsPerSlot)

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				ticker.Done()
				return
			case currentSlot := <-ticker.C():
				if s.chainStarted.IsSet() && s.cfg.initialSync.Syncing() {
					continue
				}
				valid, err := isDigestValid(digest, genesis, genRoot)
				if err != nil {
					log.Error(err)
					continue
				}
				if !valid {
					log.Warnf("Sync subnets with digest %#x are no longer valid, unsubscribing from all of them.", digest)
					// Unsubscribes from all our current subnets.
					s.reValidateSubscriptions(subscriptions, []uint64{}, topicFormat, digest)
					ticker.Done()
					return
				}

				wantedSubs := s.retrieveActiveSyncSubnets(slots.ToEpoch(currentSlot))
				// Resize as appropriate.
				s.reValidateSubscriptions(subscriptions, wantedSubs, topicFormat, digest)

				// subscribe desired aggregator subnets.
				for _, idx := range wantedSubs {
					s.subscribeSyncSubnet(subscriptions, idx, digest, validate, handle)
				}
			}
		}
	}()
}

// lookup peers for attester specific subnets.
func (s *Service) lookupAttesterSubnets(digest [4]byte, idx uint64) {
	topic := p2p.GossipTypeMapping[reflect.TypeOf(&ethpb.Attestation{})]
	subnetTopic := fmt.Sprintf(topic, digest, idx)
	if !s.validPeersExist(subnetTopic) {
		log.Infof("No peers found subscribed to attestation gossip subnet with "+
			"committee index %d. Searching network for peers subscribed to the subnet.", idx)
		// perform a search for peers with the desired committee index.
		_, err := s.cfg.p2p.FindPeersWithSubnet(s.ctx, subnetTopic, idx, flags.Get().MinimumPeersPerSubnet)
		if err != nil {
			log.WithError(err).Debug("Could not search for peers")
		}
	}
}

func (s *Service) unSubscribeFromTopic(topic string) {
	log.WithField("topic", topic).Debug("Unsubscribing from topic")
	if err := s.cfg.p2p.PubSub().UnregisterTopicValidator(topic); err != nil {
		log.WithError(err).Error("Could not unregister topic validator")
	}
	sub := s.subHandler.subForTopic(topic)
	if sub != nil {
		sub.Cancel()
	}
	s.subHandler.removeTopic(topic)
	if err := s.cfg.p2p.LeaveTopic(topic); err != nil {
		log.WithError(err).Error("Unable to leave topic")
	}
}

// find if we have peers who are subscribed to the same subnet
func (s *Service) validPeersExist(subnetTopic string) bool {
	numOfPeers := s.cfg.p2p.PubSub().ListPeers(subnetTopic + s.cfg.p2p.Encoding().ProtocolSuffix())
	return len(numOfPeers) >= flags.Get().MinimumPeersPerSubnet
}

func (s *Service) retrievePersistentSubs(currSlot types.Slot) []uint64 {
	// Persistent subscriptions from validators
	persistentSubs := s.persistentSubnetIndices()
	// Update desired topic indices for aggregator
	wantedSubs := s.aggregatorSubnetIndices(currSlot)

	// Combine subscriptions to get all requested subscriptions
	return slice.SetUint64(append(persistentSubs, wantedSubs...))
}

func (_ *Service) retrieveActiveSyncSubnets(currEpoch types.Epoch) []uint64 {
	subs := cache.SyncSubnetIDs.GetAllSubnets(currEpoch)
	return slice.SetUint64(subs)
}

// filters out required peers for the node to function, not
// pruning peers who are in our attestation subnets.
func (s *Service) filterNeededPeers(pids []peer.ID) []peer.ID {
	// Exit early if nothing to filter.
	if len(pids) == 0 {
		return pids
	}
	digest, err := s.currentForkDigest()
	if err != nil {
		log.WithError(err).Error("Could not compute fork digest")
		return pids
	}
	currSlot := s.cfg.chain.CurrentSlot()
	wantedSubs := s.retrievePersistentSubs(currSlot)
	wantedSubs = slice.SetUint64(append(wantedSubs, s.attesterSubnetIndices(currSlot)...))
	topic := p2p.GossipTypeMapping[reflect.TypeOf(&ethpb.Attestation{})]

	// Map of peers in subnets
	peerMap := make(map[peer.ID]bool)

	for _, sub := range wantedSubs {
		subnetTopic := fmt.Sprintf(topic, digest, sub) + s.cfg.p2p.Encoding().ProtocolSuffix()
		peers := s.cfg.p2p.PubSub().ListPeers(subnetTopic)
		if len(peers) > flags.Get().MinimumPeersPerSubnet {
			// In the event we have more than the minimum, we can
			// mark the remaining as viable for pruning.
			peers = peers[:flags.Get().MinimumPeersPerSubnet]
		}
		// Add peer to peer map.
		for _, p := range peers {
			// Even if the peer id has
			// already been seen we still set
			// it, as the outcome is the same.
			peerMap[p] = true
		}
	}

	// Clear out necessary peers from the peers to prune.
	newPeers := make([]peer.ID, 0, len(pids))

	for _, pid := range pids {
		if peerMap[pid] {
			continue
		}
		newPeers = append(newPeers, pid)
	}
	return newPeers
}

// Add fork digest to topic.
func (_ *Service) addDigestToTopic(topic string, digest [4]byte) string {
	if !strings.Contains(topic, "%x") {
		log.Fatal("Topic does not have appropriate formatter for digest")
	}
	return fmt.Sprintf(topic, digest)
}

// Add the digest and index to subnet topic.
func (_ *Service) addDigestAndIndexToTopic(topic string, digest [4]byte, idx uint64) string {
	if !strings.Contains(topic, "%x") {
		log.Fatal("Topic does not have appropriate formatter for digest")
	}
	return fmt.Sprintf(topic, digest, idx)
}

func (s *Service) currentForkDigest() ([4]byte, error) {
	genRoot := s.cfg.chain.GenesisValidatorsRoot()
	return forks.CreateForkDigest(s.cfg.chain.GenesisTime(), genRoot[:])
}

// Checks if the provided digest matches up with the current supposed digest.
func isDigestValid(digest [4]byte, genesis time.Time, genValRoot [32]byte) (bool, error) {
	retDigest, err := forks.CreateForkDigest(genesis, genValRoot[:])
	if err != nil {
		return false, err
	}
	isNextEpoch, err := forks.IsForkNextEpoch(genesis, genValRoot[:])
	if err != nil {
		return false, err
	}
	// In the event there is a fork the next epoch,
	// we skip the check, as we subscribe subnets an
	// epoch in advance.
	if isNextEpoch {
		return true, nil
	}
	return retDigest == digest, nil
}

func agentString(pid peer.ID, hst host.Host) string {
	agString := ""
	ok := false
	rawVersion, storeErr := hst.Peerstore().Get(pid, "AgentVersion")
	agString, ok = rawVersion.(string)
	if storeErr != nil || !ok {
		agString = ""
	}
	return agString
}

func multiAddr(pid peer.ID, stat *peers.Status) string {
	addrs, err := stat.Address(pid)
	if err != nil || addrs == nil {
		return ""
	}
	return addrs.String()
}
