package sync

import (
	"bytes"
	"context"
	"sync"
	"time"

	libp2pcore "github.com/libp2p/go-libp2p/core"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/async"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p/peers"
	p2ptypes "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p/types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd/beacon-chain/flags"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	pb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	prysmTime "gitlab.waterfall.network/waterfall/protocol/coordinator/time"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
)

// maintainPeerStatuses by infrequently polling peers for their latest status.
func (s *Service) maintainPeerStatuses() {
	// Run twice per epoch.
	interval := time.Duration(params.BeaconConfig().SlotsPerEpoch.Div(2).Mul(params.BeaconConfig().SecondsPerSlot)) * time.Second
	async.RunEvery(s.ctx, interval, func() {
		wg := new(sync.WaitGroup)
		for _, pid := range s.cfg.p2p.Peers().Connected() {
			wg.Add(1)
			go func(id peer.ID) {
				defer wg.Done()
				// If our peer status has not been updated correctly we disconnect over here
				// and set the connection state over here instead.
				if s.cfg.p2p.Host().Network().Connectedness(id) != network.Connected {
					s.cfg.p2p.Peers().SetConnectionState(id, peers.PeerDisconnecting)
					if err := s.cfg.p2p.Disconnect(id); err != nil {
						log.Errorf("Error when disconnecting with peer: %v", err)
					}
					log.WithFields(logrus.Fields{
						"func": "maintainPeerStatuses",
						"peer": id,
					}).Info("Disconnect: peer")
					s.cfg.p2p.Peers().SetConnectionState(id, peers.PeerDisconnected)
					return
				}
				// Disconnect from peers that are considered bad by any of the registered scorers.
				if s.cfg.p2p.Peers().IsBad(id) {
					log.WithFields(logrus.Fields{
						"func": "maintainPeerStatuses",
						"peer": id,
					}).Info("Disconnect: bad peer")
					s.disconnectBadPeer(s.ctx, id)
					return
				}
				// If the status hasn't been updated in the recent interval time.
				lastUpdated, err := s.cfg.p2p.Peers().ChainStateLastUpdated(id)
				if err != nil {
					// Peer has vanished; nothing to do.
					return
				}
				if prysmTime.Now().After(lastUpdated.Add(interval)) {
					if err := s.reValidatePeer(s.ctx, id); err != nil {
						log.WithField("peer", id).WithError(err).Debug("Could not revalidate peer")
						log.WithField("fn", "maintainPeerStatuses").WithField("peer", id.String()).WithError(err).Debug("Disconnect: incr BadResponses")
						s.cfg.p2p.Peers().Scorers().BadResponsesScorer().Increment(id)
					}
				}
			}(pid)
		}
		// Wait for all status checks to finish and then proceed onwards to
		// pruning excess peers.
		wg.Wait()
		peerIds := s.cfg.p2p.Peers().PeersToPrune()
		peerIds = s.filterNeededPeers(peerIds)
		for _, id := range peerIds {
			if err := s.sendGoodByeAndDisconnect(s.ctx, p2ptypes.GoodbyeCodeTooManyPeers, id); err != nil {
				log.WithField("peer", id).WithError(err).Error("Could not disconnect with peer")
			}
			log.WithFields(logrus.Fields{
				"func": "maintainPeerStatuses",
				"peer": id,
			}).Info("Disconnect: peer")
		}
	})
}

// resyncIfBehind checks periodically to see if we are in normal sync but have fallen behind our peers
// by more than an epoch, in which case we attempt a resync using the initial sync method to catch up.
func (s *Service) resyncIfBehind() {
	millisecondsPerEpoch := params.BeaconConfig().SlotsPerEpoch.Mul(1000).Mul(params.BeaconConfig().SecondsPerSlot)
	// Run sixteen times per epoch.
	interval := time.Duration(millisecondsPerEpoch/16) * time.Millisecond
	async.RunEvery(s.ctx, interval, func() {
		if s.shouldReSync() {
			syncedEpoch := slots.ToEpoch(s.cfg.chain.HeadSlot()) // nolint
			// Factor number of expected minimum sync peers, to make sure that enough peers are
			// available to resync (some peers may go away between checking non-finalized peers and
			// actual resyncing).
			highestEpoch, _ := s.cfg.p2p.Peers().BestNonFinalized(flags.Get().MinimumSyncPeers*2, syncedEpoch)
			// Check if the current node is more than 1 epoch behind.
			if highestEpoch > (syncedEpoch + 1) {
				log.WithFields(logrus.Fields{
					"currentEpoch": slots.ToEpoch(s.cfg.chain.CurrentSlot()), // nolint
					"syncedEpoch":  syncedEpoch,
					"peersEpoch":   highestEpoch,
				}).Info("Fallen behind peers; reverting to initial sync to catch up")
				numberOfTimesResyncedCounter.Inc()
				s.clearPendingSlots()
				if err := s.cfg.initialSync.Resync(); err != nil {
					log.Errorf("Could not resync chain: %v", err)
				}
			}
		}
	})
}

// shouldReSync returns true if the node is not syncing and falls behind two epochs.
func (s *Service) shouldReSync() bool {
	syncedEpoch := slots.ToEpoch(s.cfg.chain.HeadSlot())     // nolint
	currentEpoch := slots.ToEpoch(s.cfg.chain.CurrentSlot()) // nolint
	prevEpoch := types.Epoch(0)
	if currentEpoch > 1 {
		prevEpoch = currentEpoch - 1
	}
	return s.cfg.initialSync != nil && !s.cfg.initialSync.Syncing() && syncedEpoch < prevEpoch
}

// sendRPCStatusRequest for a given topic with an expected protobuf message type.
func (s *Service) sendRPCStatusRequest(ctx context.Context, id peer.ID) error {
	ctx, cancel := context.WithTimeout(ctx, respTimeout)
	defer cancel()

	headRoot, err := s.cfg.chain.HeadRoot(ctx) // nolint
	if err != nil {
		return err
	}

	forkDigest, err := s.currentForkDigest()
	if err != nil {
		return err
	}
	resp := &pb.Status{
		ForkDigest:     forkDigest[:],
		FinalizedRoot:  s.cfg.chain.FinalizedCheckpt().Root,  // nolint
		FinalizedEpoch: s.cfg.chain.FinalizedCheckpt().Epoch, // nolint
		HeadRoot:       headRoot,
		HeadSlot:       s.cfg.chain.HeadSlot(), // nolint
	}
	topic, err := p2p.TopicFromMessage(p2p.StatusMessageName, slots.ToEpoch(s.cfg.chain.CurrentSlot())) // nolint
	if err != nil {
		return err
	}
	stream, err := s.cfg.p2p.Send(ctx, resp, topic, id)
	if err != nil {
		return err
	}
	defer closeStream(stream, log)

	code, errMsg, err := ReadStatusCode(stream, s.cfg.p2p.Encoding())
	if err != nil {
		return err
	}

	if code != 0 {
		log.WithFields(logrus.Fields{
			"statusCode": code,
			"func":       "sendRPCStatusRequest",
		}).Warn("Disconnect: call IsBad: scorer incr by status code")
		log.WithField("fn", "sendRPCStatusRequest").WithField("peer", id.String()).WithError(err).Debug("Disconnect: incr BadResponses")
		s.cfg.p2p.Peers().Scorers().BadResponsesScorer().Increment(id)
		return errors.New(errMsg)
	}
	msg := &pb.Status{}
	if err := s.cfg.p2p.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return err
	}

	// If validation fails, validation error is logged, and peer status scorer will mark peer as bad.
	err = s.validateStatusMessage(ctx, msg)
	s.cfg.p2p.Peers().Scorers().PeerStatusScorer().SetPeerStatus(id, msg, err)

	if s.cfg.p2p.Peers().IsBad(id) {
		log.WithFields(logrus.Fields{
			"IsBad":                s.cfg.p2p.Peers().IsBad(id),
			"func":                 "sendRPCStatusRequest",
			"0.msg.HeadSlot":       msg.HeadSlot,
			"1.msg.FinalizedEpoch": msg.FinalizedEpoch,
		}).Warn("Disconnect: bad peer")
		s.disconnectBadPeer(s.ctx, id)
	}
	return err
}

func (s *Service) reValidatePeer(ctx context.Context, id peer.ID) error {
	s.cfg.p2p.Peers().Scorers().PeerStatusScorer().SetHeadSlot(s.cfg.chain.HeadSlot()) // nolint
	if err := s.sendRPCStatusRequest(ctx, id); err != nil {
		return err
	}
	// Do not return an error for ping requests.
	if err := s.sendPingRequest(ctx, id); err != nil {
		log.WithError(err).Debug("Could not ping peer")
	}
	return nil
}

// statusRPCHandler reads the incoming Status RPC from the peer and responds with our version of a status message.
// This handler will disconnect any peer that does not match our fork version.
func (s *Service) statusRPCHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	ctx, cancel := context.WithTimeout(ctx, ttfbTimeout)
	defer cancel()
	SetRPCStreamDeadlines(stream)
	log := log.WithField("handler", "status")
	m, ok := msg.(*pb.Status)
	if !ok {
		return errors.New("message is not type *pb.Status")
	}
	if err := s.rateLimiter.validateRequest(stream, 1); err != nil {
		return err
	}
	s.rateLimiter.add(stream, 1)

	remotePeer := stream.Conn().RemotePeer()
	if err := s.validateStatusMessage(ctx, m); err != nil {
		log.WithFields(logrus.Fields{
			"peer":  remotePeer,
			"error": err,
		}).Info("Invalid status message from peer")

		respCode := byte(0)
		switch err {
		case p2ptypes.ErrGeneric:
			respCode = responseCodeServerError
		case p2ptypes.ErrWrongForkDigestVersion:

			log.WithError(err).WithFields(logrus.Fields{
				"func": "statusRPCHandler",
				"peer": remotePeer,
			}).Info("Disconnect: bad peer status")

			// Respond with our status and disconnect with the peer.
			s.cfg.p2p.Peers().SetChainState(remotePeer, m)
			if err := s.respondWithStatus(ctx, stream); err != nil {
				return err
			}
			// Close before disconnecting, and wait for the other end to ack our response.
			closeStreamAndWait(stream, log)
			if err := s.sendGoodByeAndDisconnect(ctx, p2ptypes.GoodbyeCodeWrongNetwork, remotePeer); err != nil {
				return err
			}
			return nil
		default:
			respCode = responseCodeInvalidRequest
			s.cfg.p2p.Peers().Scorers().BadResponsesScorer().Increment(remotePeer)
			log.WithField("fn", "statusRPCHandler").WithField("peer", remotePeer.String()).WithError(err).Debug("Disconnect: incr BadResponses")
		}

		originalErr := err
		resp, err := s.generateErrorResponse(respCode, err.Error())
		if err != nil {
			log.WithError(err).Error("Could not generate a response error")
		} else if _, err := stream.Write(resp); err != nil {
			// The peer may already be ignoring us, as we disagree on fork version, so log this as debug only.
			log.WithError(err).Error("Could not write to stream")
		}

		log.WithError(err).WithFields(logrus.Fields{
			"func": "statusRPCHandler",
			"peer": remotePeer,
		}).Info("Disconnect: bad peer status")

		closeStreamAndWait(stream, log)
		if err := s.sendGoodByeAndDisconnect(ctx, p2ptypes.GoodbyeCodeGenericError, remotePeer); err != nil {
			return err
		}
		return originalErr
	}
	s.cfg.p2p.Peers().SetChainState(remotePeer, m)

	if err := s.respondWithStatus(ctx, stream); err != nil {
		return err
	}
	closeStream(stream, log)
	return nil
}

func (s *Service) respondWithStatus(ctx context.Context, stream network.Stream) error {
	headRoot, err := s.cfg.chain.HeadRoot(ctx) // nolint
	if err != nil {
		return err
	}

	forkDigest, err := s.currentForkDigest()
	if err != nil {
		return err
	}
	resp := &pb.Status{
		ForkDigest:     forkDigest[:],
		FinalizedRoot:  s.cfg.chain.FinalizedCheckpt().Root,  // nolint
		FinalizedEpoch: s.cfg.chain.FinalizedCheckpt().Epoch, // nolint
		HeadRoot:       headRoot,
		HeadSlot:       s.cfg.chain.HeadSlot(), // nolint
	}

	if _, err := stream.Write([]byte{responseCodeSuccess}); err != nil {
		log.WithError(err).Debug("Could not write to stream")
	}
	_, err = s.cfg.p2p.Encoding().EncodeWithMaxLength(stream, resp)
	return err
}

func (s *Service) validateStatusMessage(ctx context.Context, msg *pb.Status) error {
	forkDigest, err := s.currentForkDigest()
	if err != nil {
		return err
	}
	if !bytes.Equal(forkDigest[:], msg.ForkDigest) {
		return p2ptypes.ErrWrongForkDigestVersion
	}
	genesis := s.cfg.chain.GenesisTime()                   // nolint
	finalizedEpoch := s.cfg.chain.FinalizedCheckpt().Epoch // nolint
	maxEpoch := slots.EpochsSinceGenesis(genesis)
	// It would take a minimum of 2 epochs to finalize a
	// previous epoch
	maxFinalizedEpoch := types.Epoch(0)
	if maxEpoch > 2 {
		maxFinalizedEpoch = maxEpoch - 2
	}
	if msg.FinalizedEpoch > maxFinalizedEpoch {
		return p2ptypes.ErrInvalidEpoch
	}
	// Exit early if the peer's finalized epoch
	// is less than that of the remote peer's.
	if finalizedEpoch < msg.FinalizedEpoch {
		return nil
	}
	finalizedAtGenesis := msg.FinalizedEpoch == 0
	rootIsEqual := bytes.Equal(params.BeaconConfig().ZeroHash[:], msg.FinalizedRoot)
	// If peer is at genesis with the correct genesis root hash we exit.
	if finalizedAtGenesis && rootIsEqual {
		return nil
	}
	if !s.cfg.beaconDB.IsFinalizedBlock(ctx, bytesutil.ToBytes32(msg.FinalizedRoot)) {
		return p2ptypes.ErrInvalidFinalizedRoot
	}
	blk, err := s.cfg.beaconDB.Block(ctx, bytesutil.ToBytes32(msg.FinalizedRoot))
	if err != nil {
		return p2ptypes.ErrGeneric
	}
	if blk == nil || blk.IsNil() {
		return p2ptypes.ErrGeneric
	}
	if slots.ToEpoch(blk.Block().Slot()) == msg.FinalizedEpoch {
		return nil
	}

	startSlot, err := slots.EpochStart(msg.FinalizedEpoch)
	if err != nil {
		return p2ptypes.ErrGeneric
	}
	if startSlot > blk.Block().Slot() {
		childBlock, err := s.cfg.beaconDB.FinalizedChildBlock(ctx, bytesutil.ToBytes32(msg.FinalizedRoot))
		if err != nil {
			return p2ptypes.ErrGeneric
		}
		// Is a valid finalized block if no
		// other child blocks exist yet.
		if childBlock == nil || childBlock.IsNil() {
			return nil
		}
		// If child finalized block also has a smaller or
		// equal slot number we return an error.
		if startSlot >= childBlock.Block().Slot() {
			return p2ptypes.ErrInvalidEpoch
		}
		return nil
	}
	return p2ptypes.ErrInvalidEpoch
}
