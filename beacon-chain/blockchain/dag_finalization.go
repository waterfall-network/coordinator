package blockchain

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	gwatTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
	"go.opencensus.io/trace"
)

// initGwatSync initialize initial state,
// then start gwat synchronization
// and then run finalization processing
func (s *Service) initGwatSync() {
	ticker := time.NewTicker(time.Duration(params.BeaconConfig().GwatSyncIntervalMs) * time.Millisecond)
	defer func() {
		ticker.Stop()
	}()
	log.WithField("interval", fmt.Sprintf("%d", params.BeaconConfig().GwatSyncIntervalMs)).Info("Gwat sync: start ...")

	for {
		select {
		case <-s.ctx.Done():
			log.Info("Gwat sync: context closed, exiting routine")
			return
		case <-s.newHeadCh:
			continue
		case t := <-ticker.C:
			// synchronization procedure:
			var err error
			// 1. check is coordinator synchronized
			if !s.IsSynced() {
				if t.Second()%5 == 0 {
					log.WithFields(logrus.Fields{
						"currentSlot": s.CurrentSlot(),
						"headSlot":    s.headSlot(),
					}).Info("Gwat sync: coordinator is not synchronized ...")
				}
				continue
			}
			log.Info("Gwat sync: coordinator synchronization success")

			// 2. Init coordinated state
			err = s.initCoordinatedState(s.ctx)
			if err != nil {
				log.WithError(err).Warning("Gwat sync: attempt to get gwat coordinated state failed ...")
				continue
			}
			log.Info("Gwat sync: coordinated state initialization successful")

			// 3. sync gwat to current finalized checkpoint
			err = s.runGwatSynchronization(s.ctx)
			if err != nil {
				log.WithError(err).Warning("Gwat sync: attempt failed ...")
				continue
			}
			log.Info("Gwat sync: success")

			// 4. start main work process
			s.runProcessDagFinalize()
			return
		}
	}
}

// initParallelGwatSync launches parallel gwat synchronization process
// which doesn't depend on if coordinator is synced or not
func (s *Service) initParallelGwatSync(ctx context.Context) {
	log.Info("Parallel Gwat sync: start ...")

	var err error

	// 1. Check and init coordinated state
	if s.GetCachedGwatCoordinatedState() == nil {
		err = s.initCoordinatedState(ctx)
		if err != nil {
			log.WithError(err).Warning("Parallel Gwat sync: attempt to get gwat coordinated state failed ...")
			return
		}
		log.Info("Parallel Gwat sync: coordinated state initialization successful")
	}
	// 2. sync gwat to current finalized checkpoint
	err = s.runGwatSynchronization(ctx)
	if err != nil {
		log.WithError(err).Warning("Parallel Gwat sync: attempt failed ...")
		s.ResetCachedGwatCoordinatedState()
		return
	}
	log.Info("Parallel Gwat sync: success")
}

// runGwatSynchronization procedure of gwat synchronization.
func (s *Service) runGwatSynchronization(ctx context.Context) error {
	if !s.isGwatSyncing.IsSet() {
		s.isGwatSyncing.Set()
		defer func() {
			s.isGwatSyncing.UnSet()
		}()
		// skip if before genesis time
		if s.CurrentSlot() == 0 {
			return nil
		}

		gwatCheckpoint := s.GetCachedGwatCoordinatedState()
		if gwatCheckpoint == nil {
			return errNoCoordState
		}
		log.WithFields(logrus.Fields{
			"gwatCoord.Root":  fmt.Sprintf("%#x", gwatCheckpoint.Root),
			"gwatCoord.Spine": fmt.Sprintf("%#x", gwatCheckpoint.Spine),
			"gwatCoord.Epoch": gwatCheckpoint.Epoch,
		}).Info("Gwat sync: gwat coordinated state")

		cpEpoch := types.Epoch(gwatCheckpoint.Epoch)

		var syncEpoch types.Epoch
		syncSlot, err := slots.EpochStart(cpEpoch + 1)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"syncSlot": syncSlot,
				"headSlot": s.headSlot(),
				"headRoot": fmt.Sprintf("%#x", s.headRoot()),
			}).Error("Gwat sync: calc epoch start failed")
			return err
		}

		log.WithFields(logrus.Fields{
			"headSlot": s.headSlot(),
			"headRoot": fmt.Sprintf("%#x", s.headRoot()),
			"syncSlot": syncSlot,
		}).Info("Gwat sync: sync start")

		for syncSlot <= s.HeadSlot() {

			log.WithFields(logrus.Fields{
				"syncSlot": syncSlot,
				"headSlot": s.headSlot(),
			}).Debug("Gwat sync: 000")

			syncRoot := params.BeaconConfig().ZeroHash
			_, roots, err := s.cfg.BeaconDB.BlockRootsBySlot(ctx, syncSlot)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"syncSlot": syncSlot,
					"headSlot": s.headSlot(),
					"headRoot": fmt.Sprintf("%#x", s.headRoot()),
				}).Error("Gwat sync: failed 1")
				return err
			}
			if len(roots) == 0 {
				syncSlot++
				continue
			}
			if len(roots) == 1 {
				syncRoot = roots[0]
			} else {
				for _, r := range roots {
					canonical, err := s.IsCanonical(ctx, r)
					if err != nil {
						log.WithError(err).WithFields(logrus.Fields{
							"syncSlot": syncSlot,
							"headSlot": s.headSlot(),
							"headRoot": fmt.Sprintf("%#x", s.headRoot()),
						}).Error("Gwat sync: failed 2")
						return err
					}
					if canonical {
						syncRoot = r
						break
					}
				}
			}

			log.WithFields(logrus.Fields{
				"syncSlot": syncSlot,
				"headSlot": s.headSlot(),
				"syncRoot": fmt.Sprintf("%#x", syncRoot),
				"headRoot": fmt.Sprintf("%#x", s.headRoot()),
			}).Debug("Gwat sync: 111")

			if syncRoot == params.BeaconConfig().ZeroHash {
				syncSlot++
				continue
			}

			syncState, err := s.cfg.StateGen.StateByRoot(ctx, syncRoot)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"syncSlot":     syncSlot,
					"headSlot":     s.headSlot(),
					"syncRoot":     fmt.Sprintf("%#x", syncRoot),
					"headRoot":     fmt.Sprintf("%#x", s.headRoot()),
					"Prefix":       gwatCommon.HashArrayFromBytes(syncState.SpineData().Prefix),
					"Finalization": gwatCommon.HashArrayFromBytes(syncState.SpineData().Finalization),
					"CpFinalized":  gwatCommon.HashArrayFromBytes(syncState.SpineData().CpFinalized),
				}).Error("Gwat sync: failed 3")
				return err
			}

			log.WithFields(logrus.Fields{
				"epoch": slots.ToEpoch(syncState.Slot()),
				"slot":  fmt.Sprintf("%d", syncState.Slot()),
				"root":  fmt.Sprintf("%#x", syncRoot),
			}).Info("Sync state: sync")

			log.WithFields(logrus.Fields{
				"syncSlot":     syncSlot,
				"headSlot":     s.headSlot(),
				"syncRoot":     fmt.Sprintf("%#x", syncRoot),
				"headRoot":     fmt.Sprintf("%#x", s.headRoot()),
				"Prefix":       gwatCommon.HashArrayFromBytes(syncState.SpineData().Prefix),
				"Finalization": gwatCommon.HashArrayFromBytes(syncState.SpineData().Finalization),
				"CpFinalized":  gwatCommon.HashArrayFromBytes(syncState.SpineData().CpFinalized),
			}).Debug("Gwat sync: 222")

			err = s.processDagFinalization(syncState, gwatTypes.MainSync)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"syncSlot": syncSlot,
					"headSlot": s.headSlot(),
					"syncRoot": fmt.Sprintf("%#x", syncRoot),
					"headRoot": fmt.Sprintf("%#x", s.headRoot()),
				}).Error("Gwat sync: failed 4")
				return err
			}

			// sync next epoch
			syncEpoch = slots.ToEpoch(syncSlot)
			syncSlot, err = slots.EpochStart(syncEpoch + 1)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"syncSlot": syncSlot,
					"headSlot": s.headSlot(),
					"syncRoot": fmt.Sprintf("%#x", syncRoot),
					"headRoot": fmt.Sprintf("%#x", s.headRoot()),
				}).Error("Gwat sync: failed 5")
				return err
			}

			log.WithFields(logrus.Fields{
				"syncSlot": syncSlot,
				"headSlot": s.headSlot(),
				"syncRoot": fmt.Sprintf("%#x", syncRoot),
				"headRoot": fmt.Sprintf("%#x", s.headRoot()),
			}).Info("Gwat sync: main success")
		}

		//if is parallel gwat sync running
		if !s.IsSynced() {
			// set synced epoch to coordinated state to avoid duplication of finalization for next call.
			gwatCheckpoint = s.GetCachedGwatCoordinatedState()
			if gwatCheckpoint == nil {
				return errNoCoordState
			}
			if syncEpoch == 0 {
				return nil
			}
			gwatCheckpoint.Epoch = uint64(syncEpoch)
			s.CacheGwatCoordinatedState(gwatCheckpoint)
			return nil
		}

		// head sync
		headState, err := s.HeadState(ctx)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"headSlot": s.headSlot(),
				"headRoot": fmt.Sprintf("%#x", s.headRoot()),
			}).Error("Gwat sync: head state failed")
			return err
		}
		err = s.processDagFinalization(headState, gwatTypes.HeadSync)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"headSlot": s.headSlot(),
				"headRoot": fmt.Sprintf("%#x", s.headRoot()),
			}).Error("Gwat sync: head failed")
			return err
		}

		log.WithFields(logrus.Fields{
			"curSlot":  fmt.Sprintf("%#x", s.CurrentSlot()),
			"headSlot": fmt.Sprintf("%#x", s.HeadSlot()),
			"headRoot": fmt.Sprintf("%#x", s.headRoot()),
		}).Info("Gwat sync: head success")

		return nil
	} else {
		log.Info("Gwat sync: gwat syn is already in progress, skipping this call")
		return errors.New("Gwat sync is in progress")
	}
}

// runProcessDagFinalize This routine processes gwat finalization process.
func (s *Service) runProcessDagFinalize() {
	go func() {
		var headRoot []byte
		for {
			select {
			case <-s.ctx.Done():
				log.Info("Dag finalization: context done")
				return
			case newHead := <-s.newHeadCh:
				if bytes.Equal(headRoot, newHead.root[:]) {
					log.Info("Dag finalization: skip (head duplicated)")
					continue
				}

				err := s.processDagFinalization(newHead.state, gwatTypes.NoSync)
				if err != nil {
					// reset if failed
					log.WithError(err).WithFields(logrus.Fields{
						"newHead.root": fmt.Sprintf("%#x", newHead.root),
						"newHead.slot": newHead.slot,
					}).Error("Dag finalization: failed start sync sync procedure")
					go s.initGwatSync()
					return
				}

				log.WithFields(logrus.Fields{
					"StateRoot": fmt.Sprintf("%#x", newHead.block.Block().StateRoot()),
					"Slot":      newHead.state.Slot(),
					"cp.Epoch":  newHead.state.FinalizedCheckpoint().Epoch,
				}).Info("Dag finalization: success")
			}
		}
	}()
}

// createGwatSyncParam calculate and save gwat sync params
func (s *Service) saveGwatSyncState(ctx context.Context, root [32]byte) error {
	ctx, span := trace.StartSpan(ctx, "blockChain.saveGwatSyncState")
	defer span.End()

	if !s.cfg.BeaconDB.HasState(ctx, root) {
		syncState, err := s.cfg.StateGen.StateByRoot(ctx, root)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"blockRoot": fmt.Sprintf("%#x", root),
			}).Error("Save gwat sync state: get state failed")
			return err
		}
		if err = s.cfg.BeaconDB.SaveState(ctx, syncState, root); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"blockRoot": fmt.Sprintf("%#x", root),
			}).Error("Save gwat sync state: save failed")
			return err
		}
		log.WithFields(logrus.Fields{
			"slot": syncState.Slot(),
			"root": fmt.Sprintf("%#x", root),
		}).Info("Saved state in DB for sync")

		log.WithFields(logrus.Fields{
			"epoch": slots.ToEpoch(syncState.Slot()),
			"slot":  fmt.Sprintf("%d", syncState.Slot()),
			"root":  fmt.Sprintf("%#x", root),
		}).Info("Sync state: save")

	}
	return nil
}

// processDagFinalization implements dag finalization procedure.
func (s *Service) processDagFinalization(headState state.BeaconState, syncMode gwatTypes.SyncMode) error {
	ctx, span := trace.StartSpan(s.ctx, "blockChain.processDagFinalization")
	defer span.End()

	var finalizedSeq gwatCommon.HashArray

	if s.IsSynced() || syncMode == gwatTypes.MainSync {
		finParams, err := s.collectFinalizationParams(ctx, headState)
		if err != nil {
			log.WithError(err).Error("Dag finalization: get finalization params failed")
			return errors.Wrap(err, "Dag finalization: get finalization params failed")
		}
		finParams.SyncMode = syncMode
		paramCp := finParams.Checkpoint.Copy()

		log.WithFields(logrus.Fields{
			"params.Spines":    finParams.Spines,
			"params.BaseSpine": finParams.BaseSpine.Hex(),
			"cp.FinEpoch":      finParams.Checkpoint.FinEpoch,
			"cp.Epoch":         finParams.Checkpoint.Epoch,
			"cp.Root":          finParams.Checkpoint.Root.Hex(),
			"cp.Spine":         finParams.Checkpoint.Spine.Hex(),
			"params.SyncMode":  finParams.SyncMode,
		}).Info("Dag finalization: finalization params")

		finRes, err := s.cfg.ExecutionEngineCaller.ExecutionDagFinalize(ctx, finParams)
		baseSpine := finParams.BaseSpine
		finalizing := finParams.Spines
		lfSpine := finRes.LFSpine
		//fSeq := append(gwatCommon.HashArray{*baseSpine}, finalizing...)
		if err != nil || lfSpine == nil {
			log.WithError(err).WithFields(logrus.Fields{
				"params.Spines":      finParams.Spines,
				"params.BaseSpine":   finParams.BaseSpine.Hex(),
				"params.Cp.FinEpoch": finParams.Checkpoint.FinEpoch,
				"params.Cp.Epoch":    finParams.Checkpoint.Epoch,
				"params.Cp.Spine":    finParams.Checkpoint.Spine.Hex(),
				"lfSpine":            fmt.Sprintf("%#x", lfSpine),
			}).Error("Dag finalization: execution failed")
			return errors.Wrap(err, "Dag finalization: execution failed")
		}
		// cache coordinated checkpoint
		if finRes.CpEpoch != nil && finRes.CpRoot != nil {
			if paramCp.Root == *finRes.CpRoot && paramCp.Epoch == *finRes.CpEpoch {
				s.CacheGwatCoordinatedState(paramCp)
			} else {
				//get gwat matched checkpoint
				cpState, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(finRes.CpRoot.Bytes()))
				if err != nil {
					log.WithError(errors.Wrapf(err,
						"get gwat checkpoint state failed for epoch=%d root=%x", finRes.CpEpoch, finRes.CpRoot),
					).Error("Dag finalization: state error")
					return errors.Wrapf(err, "get gwat checkpoint state failed for epoch=%d root=%x", finRes.CpEpoch, finRes.CpRoot)
				}
				if cpState == nil || cpState.IsNil() {
					log.WithError(errors.Wrapf(err,
						"gwat checkpoint state not found for epoch=%d root=%x", finRes.CpEpoch, finRes.CpRoot),
					).Error("Dag finalization: state error")
					return errors.Wrapf(err, "gwat checkpoint state not found for epoch=%d root=%x", finRes.CpEpoch, finRes.CpRoot)
				}
				cpRoot := gwatCommon.BytesToHash(finRes.CpRoot.Bytes())
				cpEpoch := uint64(slots.ToEpoch(cpState.Slot()))
				finSeq := gwatCommon.HashArrayFromBytes(cpState.SpineData().Finalization)
				if len(finSeq) == 0 {
					finSeq = gwatCommon.HashArrayFromBytes(cpState.SpineData().CpFinalized)
				}
				var spine gwatCommon.Hash
				if len(finSeq) > 0 {
					spine = finSeq[len(finSeq)-1]
				}
				s.CacheGwatCoordinatedState(&gwatTypes.Checkpoint{
					Epoch: cpEpoch,
					Root:  cpRoot,
					Spine: spine,
				})
			}
		}

		log.WithFields(logrus.Fields{
			"finalized": finalizing,
			"baseSpine": fmt.Sprintf("%#x", baseSpine),
			"lfSpine":   fmt.Sprintf("%#x", lfSpine),
			//"finalizationSeq":  fSeq,
			"finalizedSeq": finalizedSeq,
			//"isFullyFinalized": fSeq.IsEqualTo(finalizedSeq),
		}).Debug("Dag finalization: finalization success")
	}

	return nil
}

// collectFinalizationParams collects params to call gwat finalization api.
func (s *Service) collectFinalizationParams(
	ctx context.Context,
	headState state.BeaconState,
) (*gwatTypes.FinalizationParams, error) {
	if headState == nil || headState.IsNil() {
		return nil, errors.New("Collect finalization params: nil head state received")
	}
	var (
		paramCpFinalized *gwatTypes.Checkpoint
		baseSpine        gwatCommon.Hash
		finalizationSeq  gwatCommon.HashArray
		err              error
	)

	// get gwat validator sync data
	valSyncData, err := s.collectValidatorSyncData(ctx, headState)
	if err != nil {
		return nil, err
	}

	// get gwat checkpoint param from state
	paramCpFinalized, err = s.getRequestGwatCheckpoint(ctx, headState)
	if err != nil {
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"paramCpFinalized.FinEpoch": paramCpFinalized.FinEpoch,
		"paramCpFinalized.Epoch":    paramCpFinalized.Epoch,
		"paramCpFinalized.Root":     fmt.Sprintf("%#x", paramCpFinalized.Root),
		"paramCpFinalized.Spine":    fmt.Sprintf("%#x", paramCpFinalized.Spine),
	}).Info("Collect finalization params: finalized checkpoint")

	cpFinalized := gwatCommon.HashArrayFromBytes(headState.SpineData().CpFinalized)
	finalization := gwatCommon.HashArrayFromBytes(headState.SpineData().Finalization)
	baseSpine = cpFinalized[0]
	finalizationSeq = append(cpFinalized, finalization...)
	if baseIx := finalizationSeq.IndexOf(baseSpine); baseIx > -1 {
		finalizationSeq = finalizationSeq[baseIx+1:]
	}

	return &gwatTypes.FinalizationParams{
		Spines:      finalizationSeq.Uniq(),
		BaseSpine:   &baseSpine,
		Checkpoint:  paramCpFinalized,
		ValSyncData: valSyncData,
	}, nil
}

// getRequestGwatCheckpoint create gwatTypes.Checkpoint param to call gwat finalization api.
func (s *Service) getRequestGwatCheckpoint(
	ctx context.Context,
	headState state.BeaconState,
) (cpFinalized *gwatTypes.Checkpoint, err error) {
	if headState == nil || headState.IsNil() {
		return cpFinalized, errors.New("Collect finalization params: nil head state received")
	}
	checkpoint := headState.FinalizedCheckpoint()
	//check cached
	if cp := s.GetCachedGwatCheckpoint(checkpoint.Root); cp != nil {
		return cp, nil
	}
	if checkpoint.Epoch == 0 {
		return s.createGenesisCoordinatedCheckpoint(ctx, slots.ToEpoch(headState.Slot()))
	}
	// create
	cpRoot := bytesutil.ToBytes32(checkpoint.Root)
	cpState, err := s.cfg.StateGen.StateByRoot(ctx, cpRoot)
	if err != nil {
		return cpFinalized, err
	}
	return &gwatTypes.Checkpoint{
		FinEpoch: uint64(slots.ToEpoch(headState.Slot())),
		Epoch:    uint64(checkpoint.Epoch),
		Root:     gwatCommon.BytesToHash(checkpoint.Root),
		Spine:    helpers.GetTerminalFinalizedSpine(cpState), // use last finalized spine
	}, nil
}

// collectValidatorSyncData collect data for ValSyncData param to call gwat finalization api.
func (s *Service) collectValidatorSyncData(ctx context.Context, st state.BeaconState) ([]*gwatTypes.ValidatorSync, error) {
	if st == nil || st.IsNil() {
		return nil, errors.New("Collect finalization params: nil head state received")
	}
	var validatorSyncData []*gwatTypes.ValidatorSync
	currentEpoch := slots.ToEpoch(st.Slot())
	vals := st.Validators()
	ffepoch := params.BeaconConfig().FarFutureEpoch

	for idx, validator := range vals {
		// activation
		if validator.ActivationEpoch < ffepoch && validator.ActivationEpoch > 0 && validator.ActivationEpoch > currentEpoch {
			validatorSyncData = append(validatorSyncData, &gwatTypes.ValidatorSync{
				OpType:     gwatTypes.Activate,
				ProcEpoch:  uint64(validator.ActivationEpoch),
				Index:      uint64(idx),
				Creator:    gwatCommon.BytesToAddress(validator.CreatorAddress),
				Amount:     nil,
				InitTxHash: gwatCommon.BytesToHash(validator.ActivationHash),
			})
			log.WithFields(logrus.Fields{
				"validator.ActivationEpoch": validator.ActivationEpoch,
				"currentEpoch":              currentEpoch,
			}).Info("activate params")
		}
		// deactivation
		if validator.ExitEpoch < ffepoch && validator.ExitEpoch > 0 && validator.ExitEpoch > currentEpoch {
			validatorSyncData = append(validatorSyncData, &gwatTypes.ValidatorSync{
				OpType:     gwatTypes.Deactivate,
				ProcEpoch:  uint64(validator.ExitEpoch),
				Index:      uint64(idx),
				Creator:    gwatCommon.BytesToAddress(validator.CreatorAddress),
				Amount:     nil,
				InitTxHash: gwatCommon.BytesToHash(validator.ExitHash),
			})

			log.WithFields(logrus.Fields{
				"validator.ExitEpoch": validator.ExitEpoch,
				"currentEpoch":        currentEpoch,
			}).Info("Exit params")
		}
	}

	// withdrawals (update balance) calculate for finalized cp
	checkpoint := st.FinalizedCheckpoint()
	cpRoot := bytesutil.ToBytes32(checkpoint.Root)
	cpState, err := s.cfg.StateGen.StateByRoot(ctx, cpRoot)
	if err != nil {
		return nil, err
	}
	minSlot, err := slots.EpochStart(cpState.FinalizedCheckpointEpoch() + 1)
	if err != nil {
		return nil, err
	}
	cpValidators := cpState.Validators()

	// collect withdrawal validator sync op
	for idx, validator := range cpValidators {
		for _, wop := range validator.WithdrawalOps {
			if wop.Slot < minSlot {
				continue
			}
			//gwei to wei
			amt := new(big.Int).Mul(new(big.Int).SetUint64(wop.Amount), new(big.Int).SetUint64(1000000000))
			vsd := &gwatTypes.ValidatorSync{
				OpType:     gwatTypes.UpdateBalance,
				ProcEpoch:  uint64(currentEpoch) + 1,
				Index:      uint64(idx),
				Creator:    gwatCommon.BytesToAddress(validator.CreatorAddress),
				Amount:     amt,
				InitTxHash: gwatCommon.BytesToHash(wop.Hash),
			}
			validatorSyncData = append(validatorSyncData, vsd)
			log.WithFields(logrus.Fields{
				"st.Slot":   st.Slot(),
				"wop.Slot":  wop.Slot,
				"valSyncOp": vsd.Print(),
			}).Info("Withdrawals: Update balance params")
		}
	}
	return validatorSyncData, nil
}

// initCoordinatedState initialize coordinated state on start up sync and finalization processing
func (s *Service) initCoordinatedState(ctx context.Context) error {
	if features.Get().EnablePassSlotInfoToGwat {
		slotInfo := &gwatTypes.SlotInfo{
			GenesisTime:    uint64(s.GenesisTime().Unix()),
			SecondsPerSlot: params.BeaconConfig().SecondsPerSlot,
			SlotsPerEpoch:  uint64(params.BeaconConfig().SlotsPerEpoch),
		}
		isSet, err := s.cfg.ExecutionEngineCaller.ExecutionDagSyncSlotInfo(s.ctx, slotInfo)
		if err != nil || !isSet {
			log.WithError(err).Warning("Gwat sync: attempt to sync slot info failed ...")
			return err
		}
		log.Info("Gwat sync: sync slot info successful")
	}
	var coordCp *gwatTypes.Checkpoint
	coordState, err := s.cfg.ExecutionEngineCaller.ExecutionDagCoordinatedState(ctx)
	if err != nil || coordState.LFSpine == nil {
		log.WithError(err).WithFields(logrus.Fields{
			"coordState": coordState,
		}).Error("Init coordinated state: get dag coordinated state failed")
		return errors.Wrap(err, "Init coordinated state: get dag coordinated state failed")
	}
	// if gwat at genesis state
	if coordState.CpRoot == nil || coordState.CpEpoch == nil || *coordState.CpRoot == (gwatCommon.Hash{}) {
		coordCp, err = s.createGenesisCoordinatedCheckpoint(ctx, 1)
		if err != nil {
			log.WithError(err).Error("Init coordinated state: create genesis state failed")
			return errors.Wrap(err, "Init coordinated state: create genesis state failed")
		}
		// check gwat genesis
		if !bytes.Equal(coordState.LFSpine.Bytes(), coordCp.Spine.Bytes()) {
			log.WithFields(logrus.Fields{
				"resivedGenesisHash": fmt.Sprintf("%#x", coordState.LFSpine),
				"gwatGenesisHash":    fmt.Sprintf("%#x", coordCp.Spine),
			}).Error("Init coordinated state: genesis spines not matched (reset to genesis spine)")
			//return errors.Wrap(err, "Init coordinated state: genesis spines not matched")
		}

		log.WithFields(logrus.Fields{
			"CpEpoch": coordCp.Epoch,
			"CpRoot":  fmt.Sprintf("%#x", coordCp.Root),
			"LFSpine": fmt.Sprintf("%#x", coordCp.Spine),
		}).Info("Init coordinated state: init by genesis")

		//cache coordinated checkpoint
		s.CacheGwatCoordinatedState(coordCp)

		//// cache last finalized spine
		//s.SetFinalizedSpinesHead(coordCp.Spine)
		return nil
	}

	// retrieve checkpoint state
	cpRoot := bytesutil.ToBytes32(coordState.CpRoot.Bytes())

	cpState, err := s.cfg.StateGen.StateByRoot(ctx, cpRoot)
	if err != nil || cpState == nil {
		log.WithError(err).WithFields(logrus.Fields{
			"cpRoot": fmt.Sprintf("%#x", cpRoot),
		}).Error("Init coordinated state: the coordinated state not found")
		return errors.Wrap(err, "Init coordinated state: the coordinated state not found")
	}

	log.WithFields(logrus.Fields{
		"cpSlot":  cpState.Slot(),
		"cpEpoch": uint64(slots.ToEpoch(cpState.Slot())),
		"cpRoot":  fmt.Sprintf("%#x", cpRoot),
	}).Debug("Init coordinated state: current result")

	finSeq := gwatCommon.HashArrayFromBytes(cpState.SpineData().Finalization)
	if len(finSeq) == 0 {
		finSeq = gwatCommon.HashArrayFromBytes(cpState.SpineData().CpFinalized)
	}
	lfspine := gwatCommon.Hash{}
	if len(finSeq) == 0 {
		log.WithFields(logrus.Fields{
			"cpSlot":   cpState.Slot(),
			"cpEpoch":  uint64(slots.ToEpoch(cpState.Slot())),
			"cpRoot":   fmt.Sprintf("%#x", cpRoot),
			"cpFinSeq": finSeq,
		}).Error("Init coordinated state: finalization empty")
		//return errors.Wrap(err, "Init coordinated state: finalization empty")
	} else {
		lfspine = finSeq[len(finSeq)-1]
	}

	// cache coordinated checkpoint
	s.CacheGwatCoordinatedState(&gwatTypes.Checkpoint{
		Epoch: uint64(slots.ToEpoch(cpState.Slot())),
		Root:  gwatCommon.BytesToHash(cpRoot[:]),
		Spine: lfspine,
	})
	return nil
}

// createGenesisCoordinatedCheckpoint create coordinated state if gwat at genesis
func (s *Service) createGenesisCoordinatedCheckpoint(ctx context.Context, cpFinEpoch types.Epoch) (*gwatTypes.Checkpoint, error) {
	genRoot, err := s.cfg.BeaconDB.GenesisBlockRoot(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get genesis root failed")
	}
	genesisSt, err := s.cfg.StateGen.StateByRoot(ctx, genRoot)
	if err != nil {
		return nil, errors.Wrap(err, "get genesis state failed")
	}
	cpEpoch := uint64(slots.ToEpoch(genesisSt.Slot()))
	cpRoot := gwatCommon.BytesToHash(genRoot[:])
	lfSpine := gwatCommon.BytesToHash(genesisSt.Eth1Data().BlockHash)
	return &gwatTypes.Checkpoint{
		Epoch:    cpEpoch,
		FinEpoch: uint64(cpFinEpoch),
		Root:     cpRoot,
		Spine:    lfSpine,
	}, nil
}
