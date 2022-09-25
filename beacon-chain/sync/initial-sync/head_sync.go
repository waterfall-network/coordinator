package initialsync

import (
	"context"
	"fmt"
	"time"

	"github.com/paulbellamy/ratecounter"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/transition"
	"github.com/prysmaticlabs/prysm/config/params"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	"github.com/prysmaticlabs/prysm/time/slots"
	"github.com/sirupsen/logrus"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
	gwatTypes "github.com/waterfall-foundation/gwat/core/types"
)

// headSync implements head-sync procedure with gwat node.
//
// Step 1 - Head Sync Ready to provide checking of readiness of both sides to head-sync main procedure,
//
// Step 2 - Head Sync (main procedure) to sync both sides from finalized epoch to head
// and make sure of the consistence of nodes.
// TODO
func (s *Service) headSync(genesis time.Time) error {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()
	transition.SkipSlotCache.Disable()
	defer transition.SkipSlotCache.Enable()

	s.counter = ratecounter.NewRateCounter(counterSeconds * time.Second)

	// Step 1 - Sync to end of finalized epoch.
	if err := s.execHeadSyncReady(ctx); err != nil {
		return err
	}

	//// Already at head, no need for 2nd phase.
	//if s.cfg.Chain.HeadSlot() == slots.Since(genesis) {
	//	return nil
	//}

	// Step 2 - sync to head from majority of peers (from no less than MinimumSyncPeers*2 peers)
	// having the same world view on non-finalized epoch.
	return s.execHeadSync(ctx, genesis)
}

// execHeadSyncReady sync from head to best known finalized epoch.
func (s *Service) execHeadSyncReady(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(params.BeaconConfig().HeadSyncReadyIntervalMs) * time.Millisecond)
	defer func() {
		ticker.Stop()
	}()

	log.WithField("HeadSyncReadyIntervalMs", fmt.Sprintf("%d", params.BeaconConfig().HeadSyncReadyIntervalMs)).Info("Head sync ready starts ...")

	for {
		checkpoint := s.cfg.Chain.CurrentJustifiedCheckpt()
		if checkpoint == nil {
			log.Error("Head sync ready: no checkpoint")
			continue
		}
		cpRoot := bytesutil.ToBytes32(checkpoint.Root)
		if cpRoot == params.BeaconConfig().ZeroHash {
			log.WithField("checkpoint.Root", fmt.Sprintf("%x", checkpoint.Root)).Error("Head sync ready: checkpoint.Root empty")
			continue
		}
		log.WithField("checkpoint.Root", checkpoint.Root).Info("Head sync ready: checkpoint")

		cpState, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(checkpoint.Root))
		if err != nil {
			log.WithField("cpState", cpState).WithError(err).Error("Head sync ready: error")
			continue
		}

		// todo ++++ INIT-SYNC ++++
		syncParam := &gwatTypes.ConsensusInfo{
			Slot: uint64(cpState.Slot()),
			//todo
			Creators:   []gwatCommon.Address{},
			Finalizing: gwatCommon.HashArrayFromBytes(cpState.Eth1Data().Finalization),
		}

		logFields := logrus.Fields{
			"checkpoint.Epoch": checkpoint.Epoch,
			"checkpoint.Root":  checkpoint.Root,
			"Slot":             syncParam.Slot,
			"Creators":         syncParam.Creators,
			"Finalizing":       syncParam.Finalizing,
		}

		log.WithFields(logFields).Info("Head sync ready: check")

		isReady, err := s.cfg.ExecutionEngineCaller.ExecutionDagHeadSyncReady(ctx, syncParam)

		log.WithField("isReady", isReady).WithError(err).Info("Head sync ready: result")
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("Head sync ready: error")
		}
		if isReady {
			log.WithFields(logFields).Info("Head sync ready: success")
			return nil
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			log.Debug("Head sync ready: context closed, exiting routine")
			return nil
		}
	}
}

// execHeadSync sync from head to best known non-finalized epoch supported by majority
// of peers (no less than MinimumSyncPeers*2 peers).
// TODO
func (s *Service) execHeadSync(ctx context.Context, genesis time.Time) error {
	queue := newBlocksQueue(ctx, &blocksQueueConfig{
		p2p:                 s.cfg.P2P,
		db:                  s.cfg.DB,
		chain:               s.cfg.Chain,
		highestExpectedSlot: slots.Since(genesis),
		mode:                modeNonConstrained,
	})
	if err := queue.start(); err != nil {
		return err
	}
	for data := range queue.fetchedData {
		s.processFetchedDataRegSync(ctx, genesis, s.cfg.Chain.HeadSlot(), data)
	}
	log.WithFields(logrus.Fields{
		"syncedSlot":  s.cfg.Chain.HeadSlot(),
		"currentSlot": slots.Since(genesis),
	}).Info("Synced to head of chain")
	if err := queue.stop(); err != nil {
		log.WithError(err).Debug("Error stopping queue")
	}

	return nil
}
