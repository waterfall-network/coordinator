package sync

import (
	"context"
	"errors"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed"
	opfeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/operation"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/monitoring/tracing"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"go.opencensus.io/trace"
)

// Clients who receive a voluntary exit on this topic MUST validate the conditions within process_voluntary_exit before
// forwarding it across the network.
func (s *Service) validateVoluntaryExit(ctx context.Context, pid peer.ID, msg *pubsub.Message) (pubsub.ValidationResult, error) {
	// Validation runs on publish (not just subscriptions), so we should approve any message from
	// ourselves.
	if pid == s.cfg.p2p.PeerID() {
		return pubsub.ValidationAccept, nil
	}

	// The head state will be too far away to validate any voluntary exit.
	if s.cfg.initialSync.Syncing() {
		return pubsub.ValidationIgnore, nil
	}

	// We should not attempt to process this message if the node is running in optimistic mode.
	// We just ignore in p2p so that the peer is not penalized.
	optimistic, err := s.cfg.chain.IsOptimistic(ctx)
	if err != nil {
		return pubsub.ValidationReject, err
	}
	if optimistic {
		return pubsub.ValidationIgnore, nil
	}

	ctx, span := trace.StartSpan(ctx, "sync.validateVoluntaryExit")
	defer span.End()

	m, err := s.decodePubsubMessage(msg)
	if err != nil {
		tracing.AnnotateError(span, err)
		return pubsub.ValidationReject, err
	}

	exit, ok := m.(*ethpb.VoluntaryExit)
	if !ok {
		return pubsub.ValidationReject, errWrongMessage
	}

	if exit == nil {
		return pubsub.ValidationReject, errNilMessage
	}
	if s.hasSeenExitIndex(exit.ValidatorIndex) {
		return pubsub.ValidationIgnore, nil
	}

	headState, err := s.cfg.chain.HeadState(ctx)
	if err != nil {
		return pubsub.ValidationIgnore, err
	}

	if uint64(exit.ValidatorIndex) >= uint64(headState.NumValidators()) {
		return pubsub.ValidationReject, errors.New("validator index is invalid")
	}
	val, err := headState.ValidatorAtIndexReadOnly(exit.ValidatorIndex)
	if err != nil {
		return pubsub.ValidationIgnore, err
	}
	if err := blocks.VerifyExitData(val, headState.Slot(), exit); err != nil {
		return pubsub.ValidationReject, err
	}

	msg.ValidatorData = exit // Used in downstream subscriber

	// Broadcast the voluntary exit on a feed to notify other services in the beacon node
	// of a received voluntary exit.
	s.cfg.operationNotifier.OperationFeed().Send(&feed.Event{
		Type: opfeed.ExitReceived,
		Data: &opfeed.ExitReceivedData{
			Exit: exit,
		},
	})

	return pubsub.ValidationAccept, nil
}

// Returns true if the node has already received a valid exit request for the validator with index `i`.
func (s *Service) hasSeenExitIndex(i types.ValidatorIndex) bool {
	s.seenExitLock.RLock()
	defer s.seenExitLock.RUnlock()
	_, seen := s.seenExitCache.Get(i)
	return seen
}

// Set exit request index `i` in seen exit request cache.
func (s *Service) setExitIndexSeen(i types.ValidatorIndex) {
	s.seenExitLock.Lock()
	defer s.seenExitLock.Unlock()
	s.seenExitCache.Add(i, true)
}
