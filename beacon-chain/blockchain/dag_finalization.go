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
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/block"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	gwatTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
	"go.opencensus.io/trace"
)

// initGwatSync initialize initial state,
// then start gwat synchronization
// and then run finalization processing
func (s *Service) initGwatSync() {
	s.isGwatSyncing = true
	ticker := time.NewTicker(time.Duration(params.BeaconConfig().GwatSyncIntervalMs) * time.Millisecond)
	defer func() {
		s.isGwatSyncing = false
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
			log.WithError(err).Warning("Gwat sync: success")

			// 4. start main work process
			s.runProcessDagFinalize()
			return
		}
	}
}

// runGwatSynchronization procedure of gwat synchronization.
func (s *Service) runGwatSynchronization(ctx context.Context) error {
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

	gwatSearchEpoch := types.Epoch(gwatCheckpoint.Epoch)

	for {
		select {
		case <-s.ctx.Done():
			log.Info("Gwat sync: context done")
			return nil
		default:
		}
		//find sync param for next checkpoint
		syncParams, err := s.searchNextGwatSyncParam(ctx, gwatSearchEpoch)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"fromEpoch": gwatSearchEpoch,
			}).Error("Gwat sync: retrieving sync param failed")
			return err
		}
		// sync reached the current epoch - sync completed
		if syncParams == nil {
			return nil
		}
		log.WithFields(logrus.Fields{
			"syncParams.FinEpoch":  syncParams.FinEpoch(),
			"syncParams.Epoch":     syncParams.Epoch(),
			"syncParams.Root":      fmt.Sprintf("%#x", syncParams.Root()),
			"syncParams.CP.Epoch":  syncParams.Param().Checkpoint.Epoch,
			"syncParams.CP.Spine":  fmt.Sprintf("%#x", syncParams.Param().Checkpoint.Spine),
			"syncParams.CP.Root":   fmt.Sprintf("%#x", syncParams.Param().Checkpoint.Root),
			"gwatCheckpoint.Epoch": gwatCheckpoint.Epoch,
		}).Info("Gwat sync: sync param retrieved")

		// sync to next checkpoint
		err = s.processGwatSync(syncParams)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"syncParams.Root":      fmt.Sprintf("%#x", syncParams.Root()),
				"syncParams.Epoch":     syncParams.Epoch(),
				"gwatCheckpoint.Epoch": gwatCheckpoint.Epoch,
			}).Error("Gwat sync: synchronization to next checkpoint failed")
			return err
		}
		gwatSearchEpoch = syncParams.Epoch()
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

				err := s.processDagFinalization(newHead.block, newHead.state)
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

// onFinCpUpdHandleGwatSyncParam calculate and save gwat sync params
func (s *Service) onFinCpUpdHandleGwatSyncParam(ctx context.Context, cp *ethpb.Checkpoint, cpFinEpoch types.Epoch) error {
	ctx, span := trace.StartSpan(ctx, "blockChain.onFinCpUpdHandleGwatSyncParam")
	defer span.End()

	//save gwat sync params
	//check gwat sync params already exists
	checkParam, err := s.cfg.BeaconDB.GwatSyncParam(ctx, cp.Epoch)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"cp.Epoch": cp.Epoch,
			"cp.Root":  fmt.Sprintf("%#x", cp.Root),
		}).Error("Save gwat sync params: db error")
		return err
	}
	// skip if already exist
	if checkParam != nil && bytes.Equal(checkParam.Root(), cp.Root) {
		return nil
	}
	//get cp state
	cpState, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(cp.Root))
	if err != nil {
		log.WithError(
			errors.Wrapf(err, "could not get checkpoint state for epoch=%d root=%x", cp.Epoch, cp.GetRoot()),
		).Error("Save gwat sync params: state error")
		return err
	}
	//get cp block
	beaconBlock, err := s.cfg.BeaconDB.Block(s.ctx, bytesutil.ToBytes32(cp.Root))
	if err != nil {
		log.WithError(
			errors.Wrapf(err, "could not get checkpoint block for epoch=%d root=%x", cp.Epoch, cp.GetRoot()),
		).Error("Save gwat sync params: block error")
		return err
	}
	// search previous checkpoint
	var prevCp *gwatTypes.Checkpoint
	prevEpoch := cp.Epoch - 1
	for {
		var prevCheckParam *wrapper.GwatSyncParam
		prevCheckParam, err = s.cfg.BeaconDB.GwatSyncParam(ctx, prevEpoch)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"cp.Epoch": cp.Epoch,
				"cp.Root":  fmt.Sprintf("%#x", cp.Root),
			}).Error("Save gwat sync params: previous checkpoint error")
			return err
		}
		if prevCheckParam != nil {
			prevSpines := prevCheckParam.Param().Spines
			lastSpine := prevSpines[len(prevSpines)-1]
			prevCp = &gwatTypes.Checkpoint{
				Epoch:    uint64(prevCheckParam.Epoch()),
				FinEpoch: uint64(prevCheckParam.FinEpoch()),
				Root:     gwatCommon.BytesToHash(prevCheckParam.Root()),
				Spine:    lastSpine,
			}
			break
		}
		prevEpoch--
		if prevEpoch == 0 {
			prevCp, err = s.createGenesisCoordinatedCheckpoint(ctx)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"cpFinEpoch": cpFinEpoch,
					"cp.Epoch":   cp.Epoch,
					"cp.Root":    fmt.Sprintf("%#x", cp.Root),
				}).Error("Save gwat sync params: create genesis checkpoint failed")
				return err
			}
			break
		}
	}

	log.WithFields(logrus.Fields{
		"cpFinEpoch":      cpFinEpoch,
		"cp.Epoch":        cp.Epoch,
		"cp.Root":         fmt.Sprintf("%#x", cp.Root),
		"prevCp.FinEpoch": prevCp.FinEpoch,
		"prevCp.Epoch":    prevCp.Epoch,
		"prevCp.Root":     fmt.Sprintf("%#x", prevCp.Root),
		"prevCp.Spine":    fmt.Sprintf("%#x", prevCp.Spine),
	}).Debug("Save gwat sync params:: previous checkpoint found")

	// collect finalization params
	finParams, err := s.collectGwatSyncParams(ctx, beaconBlock, cpState, prevCp)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"cp.Epoch":   cp.Epoch,
			"state.Slot": cpState.Slot(),
		}).Error("Save gwat sync params: calculate gwat sync param failed")
		return err
	}
	log.WithFields(logrus.Fields{
		"cpFinEpoch":              cpFinEpoch,
		"cp.Epoch":                cp.Epoch,
		"cp.Root":                 fmt.Sprintf("%#x", cp.Root),
		"gwatSyncParam.Spines":    finParams.Spines,
		"gwatSyncParam.BaseSpine": fmt.Sprintf("%#x", finParams.BaseSpine),
		"finParams.CP.Spine":      fmt.Sprintf("%#x", finParams.Checkpoint.Spine),
		"finParams.CP.Root":       fmt.Sprintf("%#x", finParams.Checkpoint.Root),
		"finParams.CP.Epoch":      finParams.Checkpoint.Epoch,
		"finParams.CP.FinEpoch":   finParams.Checkpoint.FinEpoch,
	}).Info("Save gwat sync params:: gwat sync param calculated")

	//Save gwat sync param
	gsp := wrapper.NewGwatSyncParam(cp, finParams, cpFinEpoch)
	err = s.cfg.BeaconDB.SaveGwatSyncParam(ctx, *gsp)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"cpFinEpoch": cpFinEpoch,
			"cp.Epoch":   cp.Epoch,
			"state.Slot": cpState.Slot(),
		}).Error("Save gwat sync params: save param error")
		return err
	}
	return nil
}

// processGwatSync implements gwat sync procedure.
func (s *Service) processGwatSync(gsp *wrapper.GwatSyncParam) error {
	ctx, span := trace.StartSpan(s.ctx, "blockChain.processGwatSync")
	defer span.End()

	var finalizedSeq gwatCommon.HashArray
	syncParams := gsp.Param()
	paramCp := syncParams.Checkpoint.Copy()

	log.WithFields(logrus.Fields{
		"syncParams.Epoch":    gsp.Epoch(),
		"syncParams.Root":     fmt.Sprintf("%#x", gsp.Root()),
		"syncParams.CP.Epoch": gsp.Param().Checkpoint.Epoch,
		"syncParams.CP.Spine": fmt.Sprintf("%#x", gsp.Param().Checkpoint.Spine),
		"syncParams.CP.Root":  fmt.Sprintf("%#x", gsp.Param().Checkpoint.Root),
	}).Info("Gwat sync: gwat sync param")

	finRes, err := s.cfg.ExecutionEngineCaller.ExecutionDagFinalize(ctx, syncParams)
	baseSpine := syncParams.BaseSpine
	finalizing := syncParams.Spines
	lfSpine := finRes.LFSpine
	fSeq := append(gwatCommon.HashArray{*baseSpine}, finalizing...)
	if err != nil || lfSpine == nil {
		log.WithError(err).WithFields(logrus.Fields{
			"syncParams.Spines":           syncParams.Spines,
			"syncParams.BaseSpine":        syncParams.BaseSpine.Hex(),
			"checkpoint":                  syncParams.Checkpoint.Epoch,
			"syncParams.Checkpoint.Spine": syncParams.Checkpoint.Spine.Hex(),
			"lfSpine":                     fmt.Sprintf("%#x", lfSpine),
		}).Warn("Gwat sync: finalization failed")
		return errors.Wrap(err, "Gwat sync: execution failed")
	}

	log.WithFields(logrus.Fields{
		"syncParams.Spines":    syncParams.Spines,
		"syncParams.BaseSpine": syncParams.BaseSpine.Hex(),
		"checkpoint":           syncParams.Checkpoint.Epoch,
		"syncParams.CP.Spine":  syncParams.Checkpoint.Spine.Hex(),
		"lfSpine":              fmt.Sprintf("%#x", lfSpine),
		"finRes":               finRes,
		"isNoResCp":            finRes.CpEpoch != nil && finRes.CpRoot != nil,
	}).Info("Gwat sync: execution result")

	// cache coordinated checkpoint
	if finRes.CpEpoch != nil && finRes.CpRoot != nil {
		if paramCp.Root == *finRes.CpRoot && paramCp.Epoch == *finRes.CpEpoch {
			s.CacheGwatCoordinatedState(paramCp)
		} else {
			//get gwat matched checkpoint
			cpState, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(finRes.CpRoot.Bytes()))
			if err != nil {
				log.WithError(
					errors.Wrapf(err, "get gwat checkpoint state failed for epoch=%d root=%x", finRes.CpEpoch, finRes.CpRoot),
				).Error("Gwat sync: err")
				return errors.Wrapf(err, "get gwat checkpoint state failed for epoch=%d root=%x", finRes.CpEpoch, finRes.CpRoot)
			}
			if cpState == nil || cpState.IsNil() {
				log.WithError(
					errors.Wrapf(err, "gwat checkpoint state not found for epoch=%d root=%x", finRes.CpEpoch, finRes.CpRoot),
				).Error("Gwat sync: err")
				return errors.Wrapf(err, "gwat checkpoint state not found for epoch=%d root=%x", finRes.CpEpoch, finRes.CpRoot)
			}
			cpRoot := gwatCommon.BytesToHash(finRes.CpRoot.Bytes())
			cpEpoch := uint64(slots.ToEpoch(cpState.Slot()))
			finSeq := gwatCommon.HashArrayFromBytes(cpState.SpineData().Finalization)
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

	//for _, h := range fSeq {
	//	finalizedSeq = append(finalizedSeq, h)
	//	if h == *lfSpine {
	//		break
	//	}
	//}
	//
	//if len(finalizedSeq) == 0 {
	//	err = errors.New("Gwat sync: lf spine is invalid")
	//	log.WithError(err).WithFields(logrus.Fields{
	//		"finalizationSeq": fSeq,
	//		"lfSpine":         fmt.Sprintf("%#x", lfSpine),
	//		"isValid":         fSeq.Has(*lfSpine),
	//	}).Error("Gwat sync: failed")
	//	return errors.Wrap(err, "Gwat sync: failed")
	//}

	log.WithFields(logrus.Fields{
		"finalized":        finalizing,
		"baseSpine":        baseSpine.Hex(),
		"lfSpine":          lfSpine.Hex(),
		"finalizationSeq":  fSeq,
		"finalizedSeq":     finalizedSeq,
		"isFullyFinalized": fSeq.IsEqualTo(finalizedSeq),
	}).Info("Gwat sync: finalization success")

	////update checkpoint of FinalizedSpines cache
	//s.SetFinalizedSpinesCheckpoint(gsp.Param().Checkpoint.Spine)

	////update FinalizedSpines cache
	//s.AddFinalizedSpines(finalizedSeq)

	return nil
}

// processDagFinalization implements dag finalization procedure.
func (s *Service) processDagFinalization(headBlock block.SignedBeaconBlock, headState state.BeaconState) error {
	ctx, span := trace.StartSpan(s.ctx, "blockChain.processDagFinalization")
	defer span.End()

	var finalizedSeq gwatCommon.HashArray

	if s.IsSynced() {
		coordState := s.GetCachedGwatCoordinatedState()
		if coordState == nil {
			return errNoCoordState
		}
		finParams, err := s.collectFinalizationParams(ctx, headBlock, headState, coordState)
		if err != nil {
			log.WithError(err).Error("Dag finalization: get finalization params failed")
			return errors.Wrap(err, "Dag finalization: get finalization params failed")
		}
		paramCp := finParams.Checkpoint.Copy()

		log.WithFields(logrus.Fields{
			"params.Spines":    finParams.Spines,
			"params.BaseSpine": finParams.BaseSpine.Hex(),
			"cp.FinEpoch":      finParams.Checkpoint.FinEpoch,
			"cp.Epch":          finParams.Checkpoint.Epoch,
			"cp.Root":          finParams.Checkpoint.Root.Hex(),
			"cp.Spine":         finParams.Checkpoint.Spine.Hex(),
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

		//for _, h := range fSeq {
		//	finalizedSeq = append(finalizedSeq, h)
		//	if h == *lfSpine {
		//		break
		//	}
		//}

		//if len(finalizedSeq) == 0 {
		//	err = errors.New("lf spine is invalid")
		//	log.WithError(err).WithFields(logrus.Fields{
		//		"finalizationSeq": fSeq,
		//		"lfSpine":         lfSpine.Hex(),
		//		"isValid":         fSeq.Has(*lfSpine),
		//	}).Error("Dag finalization: finalization failed")
		//	return errors.Wrap(err, "Dag finalization: gwat finalization failed")
		//}

		log.WithFields(logrus.Fields{
			"finalized": finalizing,
			"baseSpine": fmt.Sprintf("%#x", baseSpine),
			"lfSpine":   fmt.Sprintf("%#x", lfSpine),
			//"finalizationSeq":  fSeq,
			"finalizedSeq": finalizedSeq,
			//"isFullyFinalized": fSeq.IsEqualTo(finalizedSeq),
		}).Debug("Dag finalization: finalization success")
	}

	////update checkpoint of FinalizedSpines cache
	//cpFin := headState.FinalizedCheckpoint()
	//cpState, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(cpFin.Root))
	//if err != nil {
	//	log.WithError(errors.Wrapf(err,
	//		"Cache finalized spines: could not get checkpoint state for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot(),
	//	)).Error("Dag finalization: state error")
	//	return errors.Wrapf(err, "Cache finalized spines: could not get checkpoint state for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot())
	//}
	//if cpState == nil || cpState.IsNil() {
	//	log.WithError(errors.Wrapf(err,
	//		"Cache finalized spines: checkpoint's state not found for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot()),
	//	).Error("Dag finalization: state error")
	//	return errors.Wrapf(err, "Cache finalized spines: checkpoint's state not found for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot())
	//}
	//s.SetFinalizedSpinesCheckpoint(helpers.GetTerminalFinalizedSpine(cpState))

	////update FinalizedSpines cache
	//s.AddFinalizedSpines(finalizedSeq)

	return nil
}

// collectFinalizationParams collects params to call gwat finalization api.
func (s *Service) collectFinalizationParams(
	ctx context.Context,
	headBlock block.SignedBeaconBlock,
	headState state.BeaconState,
	coordCheckpoint *gwatTypes.Checkpoint,
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
	valSyncData, err := collectValidatorSyncData(headState)
	if err != nil {
		return nil, err
	}

	// get gwat checkpoint param from state
	paramCpFinalized, err = s.getRequestGwatCheckpoint(ctx, headState)
	if err != nil {
		return nil, err
	}

	cpSlot, err := slots.EpochStart(types.Epoch(coordCheckpoint.Epoch))
	if err != nil {
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"cpSlot":                    cpSlot,
		"paramCpFinalized.FinEpoch": paramCpFinalized.FinEpoch,
		"paramCpFinalized.Epoch":    paramCpFinalized.Epoch,
		"paramCpFinalized.Root":     fmt.Sprintf("%#x", paramCpFinalized.Root),
		"paramCpFinalized.Spine":    fmt.Sprintf("%#x", paramCpFinalized.Spine),
	}).Debug("Collect finalization params: finalized checkpoint")

	//finalizedSpines := s.GetFinalizedSpines()
	//var currRoot [32]byte
	//currState := headState
	//currBlock := headBlock
	//for {
	//	log.WithFields(logrus.Fields{
	//		"currState.Slot":      currState.Slot(),
	//		"currBlock.Slot":      currBlock.Block().Slot(),
	//		"currBlock.StateRoot": fmt.Sprintf("%#x", currBlock.Block().StateRoot()),
	//		"cpSlot":              cpSlot,
	//	}).Debug("Collect finalization params: collect data")
	//
	//	currFinalization := gwatCommon.HashArrayFromBytes(currState.SpineData().Finalization)
	//	intersect := finalizedSpines.SequenceIntersection(currFinalization).Uniq()
	//	if len(intersect) == 0 {
	//		baseSpine = helpers.GetTerminalFinalizedSpine(currState)
	//		finalizationSeq = append(currFinalization, finalizationSeq...)
	//
	//		log.WithFields(logrus.Fields{
	//			"intersect":           intersect,
	//			"currFinalization":    currFinalization,
	//			"finalizationSeq":     finalizationSeq,
	//			"currBlock.StateRoot": fmt.Sprintf("%#x", currBlock.Block().StateRoot()),
	//			"baseSpine":           fmt.Sprintf("%#x", baseSpine),
	//			"cpSlot":              cpSlot,
	//		}).Debug("Collect finalization params: collect data: intersect empty")
	//
	//	} else {
	//		baseSpine = intersect[len(intersect)-1]
	//		add := false
	//
	//		log.WithFields(logrus.Fields{
	//			"intersect":        intersect,
	//			"currFinalization": currFinalization,
	//			"finalizationSeq":  finalizationSeq,
	//			"baseSpine":        fmt.Sprintf("%#x", baseSpine),
	//			"cpSlot":           cpSlot,
	//		}).Debug("Collect finalization params: collect data: intersect not empty")
	//
	//		for _, h := range currFinalization {
	//			if add {
	//				finalizationSeq = append(gwatCommon.HashArray{h}, finalizationSeq...)
	//			}
	//			if h == baseSpine {
	//				add = true
	//			}
	//		}
	//		break
	//	}
	//	//update FinalizedSpines cache
	//	s.SetFinalizedSpinesHead(baseSpine)
	//	//set next block root as current
	//	currRoot = bytesutil.ToBytes32(currBlock.Block().ParentRoot())
	//	if currRoot == params.BeaconConfig().ZeroHash {
	//		return &gwatTypes.FinalizationParams{
	//			Spines:      finalizationSeq.Uniq(),
	//			BaseSpine:   &baseSpine,
	//			Checkpoint:  paramCpFinalized,
	//			ValSyncData: valSyncData,
	//		}, nil
	//	}
	//	//set next block as current
	//	currBlock, err = s.cfg.BeaconDB.Block(s.ctx, currRoot)
	//	if err != nil {
	//		return nil, err
	//	}
	//	//set next state as current
	//	currState, err = s.cfg.StateGen.StateByRoot(ctx, currRoot)
	//	if err != nil {
	//		err = errors.Wrapf(err, "could not get parent state for root=%x", currRoot)
	//		log.WithError(err).Error("Collect finalization params: eror")
	//		return nil, err
	//	}
	//	if currState == nil || currState.IsNil() {
	//		err = errors.Errorf("retrieved nil parent state for root=%x", currRoot)
	//		log.WithError(err).Error("Collect finalization params: eror")
	//		return nil, err
	//	}
	//
	//	log.WithFields(logrus.Fields{
	//		"currState.Slot":       currState.Slot(),
	//		"currBlock.Slot":       currBlock.Block().Slot(),
	//		"currRoot":             fmt.Sprintf("%#x", currRoot),
	//		"currBlock.StateRoot":  fmt.Sprintf("%#x", currBlock.Block().StateRoot()),
	//		"coordCheckpoint.Root": fmt.Sprintf("%#x", coordCheckpoint.Root),
	//		"cpSlot":               cpSlot,
	//		"isErr-Slot":           currBlock.Block().Slot() < cpSlot,
	//		"isCp":                 coordCheckpoint.Root == currRoot,
	//	}).Debug("Collect finalization params: current result")
	//
	//	if coordCheckpoint.Root == currRoot {
	//		//update FinalizedSpines cache
	//		s.ResetFinalizedSpines()
	//		s.SetFinalizedSpinesHead(coordCheckpoint.Spine)
	//		break
	//	}
	//
	//	// if reach finalized checkpoint slot
	//	if currBlock.Block().Slot() < cpSlot {
	//		err = errors.New("Collect finalization params: failed")
	//		log.WithError(err).WithFields(logrus.Fields{
	//			"currBlock.Block().Slot()": currBlock.Block().Slot(),
	//			"cpSlot":                   cpSlot,
	//		}).Error("Collect finalization params: failed")
	//		return nil, err
	//	}
	//}

	finalizationSeq = gwatCommon.HashArrayFromBytes(headState.SpineData().Finalization)
	finalizationSeq = append(finalizationSeq, gwatCommon.HashArrayFromBytes(headState.SpineData().Prefix)...)
	baseSpine = helpers.GetTerminalFinalizedSpine(headState)

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
		return s.createGenesisCoordinatedCheckpoint(ctx)
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
func collectValidatorSyncData(st state.BeaconState) ([]*gwatTypes.ValidatorSync, error) {
	var validatorSyncData []*gwatTypes.ValidatorSync
	currentEpoch := slots.ToEpoch(st.Slot())
	vals := st.Validators()
	ffepoch := params.BeaconConfig().FarFutureEpoch

	for idx, validator := range vals {
		// activation
		if validator.ActivationEpoch < ffepoch && validator.ActivationEpoch > 0 && validator.ActivationEpoch > currentEpoch {
			validatorSyncData = append(validatorSyncData, &gwatTypes.ValidatorSync{
				OpType:    gwatTypes.Activate,
				ProcEpoch: uint64(validator.ActivationEpoch),
				Index:     uint64(idx),
				Creator:   gwatCommon.BytesToAddress(validator.CreatorAddress),
				Amount:    nil,
			})
			log.WithFields(logrus.Fields{
				"validator.ActivationEpoch": validator.ActivationEpoch,
				"currentEpoch":              currentEpoch,
			}).Info("activate params")
		}
		// exit
		if validator.ExitEpoch < ffepoch && validator.ExitEpoch > 0 && validator.ExitEpoch > currentEpoch {
			validatorSyncData = append(validatorSyncData, &gwatTypes.ValidatorSync{
				OpType:    gwatTypes.Deactivate,
				ProcEpoch: uint64(validator.ExitEpoch),
				Index:     uint64(idx),
				Creator:   gwatCommon.BytesToAddress(validator.CreatorAddress),
				Amount:    nil,
			})

			log.WithFields(logrus.Fields{
				"validator.ExitEpoch": validator.ExitEpoch,
				"currentEpoch":        currentEpoch,
			}).Info("Exit params")
		}
		// withdrawal
		isWitdrowalPeriod, err := helpers.IsWithdrawBalanceLocked(st, types.ValidatorIndex(idx))
		if err != nil {
			return nil, err
		}
		if isWitdrowalPeriod {
			balAtIdx, err := st.BalanceAtIndex(types.ValidatorIndex(idx))
			if err != nil {
				return nil, err
			}
			//gwei to wei
			amt := new(big.Int).Mul(new(big.Int).SetUint64(balAtIdx), new(big.Int).SetUint64(1000000000))

			vsd := &gwatTypes.ValidatorSync{
				OpType:    gwatTypes.UpdateBalance,
				ProcEpoch: uint64(validator.WithdrawableEpoch) - 1,
				Index:     uint64(idx),
				Creator:   gwatCommon.BytesToAddress(validator.CreatorAddress),
				Amount:    amt,
			}

			validatorSyncData = append(validatorSyncData, vsd)

			log.WithFields(logrus.Fields{
				"currState.Slot": st.Slot(),
				"OpType":         vsd.OpType,
				"ProcEpoch":      vsd.ProcEpoch,
				"Index":          vsd.Index,
				"Creator":        fmt.Sprintf("%#x", vsd.Creator),
				"Amount":         vsd.Amount.String(),
			}).Info("Update balance params")
		}
	}
	return validatorSyncData, nil
}

// initCoordinatedState initialize coordinated state on start up sync and finalization processing
func (s *Service) initCoordinatedState(ctx context.Context) error {
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
		coordCp, err = s.createGenesisCoordinatedCheckpoint(ctx)
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
		log.WithFields(logrus.Fields{
			"cpSlot":   cpState.Slot(),
			"cpEpoch":  uint64(slots.ToEpoch(cpState.Slot())),
			"cpRoot":   fmt.Sprintf("%#x", cpRoot),
			"cpFinSeq": finSeq,
		}).Error("Init coordinated state: finalization empty")
		return errors.Wrap(err, "Init coordinated state: finalization empty")
	}

	// cache coordinated checkpoint
	s.CacheGwatCoordinatedState(&gwatTypes.Checkpoint{
		Epoch: uint64(slots.ToEpoch(cpState.Slot())),
		Root:  gwatCommon.BytesToHash(cpRoot[:]),
		Spine: finSeq[len(finSeq)-1],
	})
	return nil
}

// createGenesisCoordinatedCheckpoint create coordinated state if gwat at genesis
func (s *Service) createGenesisCoordinatedCheckpoint(ctx context.Context) (*gwatTypes.Checkpoint, error) {
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
		FinEpoch: cpEpoch,
		Root:     cpRoot,
		Spine:    lfSpine,
	}, nil
}

// collectGwatSyncParams collects params to call gwat sync.
func (s *Service) collectGwatSyncParams(
	ctx context.Context,
	headBlock block.SignedBeaconBlock,
	headState state.BeaconState,
	coordCheckpoint *gwatTypes.Checkpoint,
) (*gwatTypes.FinalizationParams, error) {
	if headState == nil || headState.IsNil() {
		return nil, errors.New("Collect finalization params: nil head state received")
	}
	var (
		paramCpFinalized *gwatTypes.Checkpoint
		baseSpine        gwatCommon.Hash
		finalizationSeq  gwatCommon.HashArray
		cpState          state.BeaconState
		err              error
	)

	cpRoot := headState.CurrentJustifiedCheckpoint().Root
	cpState, err = s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(cpRoot))
	if err != nil {
		err = errors.Wrapf(err, "could not get parent state for root=%x", cpRoot)
		log.WithError(err).Error("Collect gwat sync params: failed")
		return nil, err
	}

	valSyncData, err := collectValidatorSyncData(cpState)
	if err != nil {
		return nil, err
	}

	paramCpFinalized = coordCheckpoint
	baseSpine = coordCheckpoint.Spine

	cpSlot, err := slots.EpochStart(types.Epoch(coordCheckpoint.Epoch))
	if err != nil {
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"cpSlot":                 cpSlot,
		"paramCpFinalized.Epoch": paramCpFinalized.Epoch,
		"paramCpFinalized.Root":  fmt.Sprintf("%#x", paramCpFinalized.Root),
		"paramCpFinalized.Spine": fmt.Sprintf("%#x", paramCpFinalized.Spine),
	}).Debug("Collect gwat sync params: checkpoint")

	var finalizedSpines = gwatCommon.HashArray{}
	var currRoot [32]byte
	currState := headState
	currBlock := headBlock
	for {
		log.WithFields(logrus.Fields{
			"currState.Slot":      currState.Slot(),
			"currBlock.Slot":      currBlock.Block().Slot(),
			"currBlock.StateRoot": fmt.Sprintf("%#x", currBlock.Block().StateRoot()),
			"cpSlot":              cpSlot,
		}).Debug("Collect gwat sync params: cycle start")

		currFinalization := gwatCommon.HashArrayFromBytes(currState.SpineData().Finalization)
		intersect := finalizedSpines.SequenceIntersection(currFinalization).Uniq()
		if len(intersect) == 0 {
			log.WithFields(logrus.Fields{
				"intersect":           intersect,
				"currFinalization":    currFinalization,
				"finalizationSeq":     finalizationSeq,
				"currBlock.StateRoot": fmt.Sprintf("%#x", currBlock.Block().StateRoot()),
				"baseSpine":           fmt.Sprintf("%#x", baseSpine),
				"cpSlot":              cpSlot,
			}).Debug("Collect gwat sync params: intersect empty")

			finalizationSeq = append(currFinalization, finalizationSeq...)
		} else {
			log.WithFields(logrus.Fields{
				"intersect":        intersect,
				"currFinalization": currFinalization,
				"finalizationSeq":  finalizationSeq,
				"baseSpine":        fmt.Sprintf("%#x", baseSpine),
				"cpSlot":           cpSlot,
			}).Debug("Collect gwat sync params: intersect not empty")

			add := false
			for _, h := range currFinalization {
				if add {
					//finalizationSeq = append(finalizationSeq, h)
					finalizationSeq = append(gwatCommon.HashArray{h}, finalizationSeq...)
				}
				if h == baseSpine {
					add = true
				}
			}
			break
		}
		//set next block root as current
		currRoot = bytesutil.ToBytes32(currBlock.Block().ParentRoot())
		if currRoot == params.BeaconConfig().ZeroHash {
			return &gwatTypes.FinalizationParams{
				Spines:      finalizationSeq.Uniq(),
				BaseSpine:   &baseSpine,
				Checkpoint:  paramCpFinalized,
				ValSyncData: valSyncData,
			}, nil
		}
		//set next block as current
		currBlock, err = s.cfg.BeaconDB.Block(s.ctx, currRoot)
		if err != nil {
			return nil, err
		}
		//set next state as current
		currState, err = s.cfg.StateGen.StateByRoot(ctx, currRoot)
		if err != nil {
			err = errors.Wrapf(err, "could not get parent state for root=%x", currRoot)
			log.WithError(err).Error("Collect gwat sync params: failed")
			return nil, err
		}
		if currState == nil || currState.IsNil() {
			err = errors.Errorf("retrieved nil parent state for root=%x", currRoot)
			log.WithError(err).Error("Collect gwat sync params: failed")
			return nil, err
		}

		log.WithFields(logrus.Fields{
			"currState.Slot":       currState.Slot(),
			"currBlock.Slot":       currBlock.Block().Slot(),
			"currRoot":             fmt.Sprintf("%#x", currRoot),
			"currBlock.StateRoot":  fmt.Sprintf("%#x", currBlock.Block().StateRoot()),
			"coordCheckpoint.Root": fmt.Sprintf("%#x", coordCheckpoint.Root),
			"cpSlot":               cpSlot,
			"isErr-Slot":           currBlock.Block().Slot() < cpSlot,
			"isCp":                 coordCheckpoint.Root == currRoot,
		}).Debug("Collect gwat sync params: cycle end")

		if coordCheckpoint.Root == currRoot {
			break
		}

		// if reach finalized checkpoint slot
		if currBlock.Block().Slot() < cpSlot {
			err = errors.New("Collect finalization params: failed")
			log.WithError(err).WithFields(logrus.Fields{
				"currBlock.Slot": currBlock.Block().Slot(),
				"cpSlot":         cpSlot,
			}).Error("Collect gwat sync params: failed")
			return nil, err
		}
	}
	return &gwatTypes.FinalizationParams{
		Spines:      finalizationSeq.Uniq(),
		BaseSpine:   &baseSpine,
		Checkpoint:  paramCpFinalized,
		ValSyncData: valSyncData,
	}, nil
}

// searchNextGwatSyncParam procedure to find next gwat synchronization param
// starting from passed gwatEpoch.
func (s *Service) searchNextGwatSyncParam(ctx context.Context, gwatEpoch types.Epoch) (*wrapper.GwatSyncParam, error) {
	nextEpoch := gwatEpoch
	for {
		nextEpoch++
		syncParam, err := s.cfg.BeaconDB.GwatSyncParam(ctx, nextEpoch)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"gwatEpoch": gwatEpoch,
				"nextEpoch": nextEpoch,
			}).Error("Gwat sync: search next gwat sync param failed")
			return nil, err
		}
		if syncParam != nil {
			return syncParam, nil
		}
		// current epoch reached
		if nextEpoch >= slots.ToEpoch(s.CurrentSlot()) {
			return nil, nil
		}
	}
}
