package blockchain

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/block"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	gwatTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
	"go.opencensus.io/trace"
)

// This routine processes gwat finalization process.
func (s *Service) spawnProcessDagFinalize() {
	go func() {
		for {
			var headRoot []byte
			select {
			case <-s.ctx.Done():
				log.Info("dag finalization: context done")
				return
			case newHead := <-s.newHeadCh:
				if bytes.Equal(headRoot, newHead.root[:]) {
					log.Info("dag finalization: skip (head duplicated)")
					continue
				}
				headRoot = bytesutil.SafeCopyBytes(newHead.root[:])
				var (
					lastFinalized gwatCommon.Hash
					lastSpine     gwatCommon.Hash
				)
				finalizedSpines := s.GetFinalizedSpines()
				if len(finalizedSpines) > 0 {
					lastFinalized = finalizedSpines[len(finalizedSpines)-1]
				}
				finalization := gwatCommon.HashArrayFromBytes(newHead.state.Eth1Data().Finalization)
				if len(finalization) > 0 {
					lastSpine = finalization[len(finalization)-1]
				}
				if lastFinalized == lastSpine {
					log.Info("dag finalization: skip (no updates)")
					continue
				}

				err := s.processDagFinalization(newHead.block, newHead.state)
				if err != nil {
					// reset if failed
					headRoot = bytesutil.SafeCopyBytes(params.BeaconConfig().ZeroHash[:])
					log.WithError(err).WithFields(logrus.Fields{
						"newHead.root": fmt.Sprintf("%#x", newHead.root),
						"newHead.slot": newHead.slot,
					}).Error("dag finalization: failed")
				}
			}
		}
	}()
}

// processDagFinalization implements dag finalization procedure.
func (s *Service) processDagFinalization(headBlock block.SignedBeaconBlock, headState state.BeaconState) error {
	ctx, span := trace.StartSpan(s.ctx, "blockChain.dagFinalizeProcess")
	defer span.End()

	var finalizedSeq gwatCommon.HashArray

	if !s.isSync() {
		cp := headState.CurrentJustifiedCheckpoint()
		cpState, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(cp.Root))
		if err != nil {
			log.WithError(errors.Wrapf(err, "could not get checkpoint state for epoch=%d root=%x", cp.Epoch, cp.GetRoot())).Error("dag finalization")
			return errors.Wrapf(err, "could not get checkpoint state for epoch=%d root=%x", cp.Epoch, cp.GetRoot())
		}
		if cpState == nil || cpState.IsNil() {
			log.WithError(errors.Wrapf(err, "checkpoint's state not found for epoch=%d root=%x", cp.Epoch, cp.GetRoot())).Error("dag finalization")
			return errors.Wrapf(err, "checkpoint's state not found for epoch=%d root=%x", cp.Epoch, cp.GetRoot())
		}

		cpFin := gwatCommon.HashArrayFromBytes(cpState.Eth1Data().Finalization)
		headFin := gwatCommon.HashArrayFromBytes(headState.Eth1Data().Finalization)
		skip := headFin.IsEqualTo(cpFin)
		if !skip {
			params, err := s.collectFinalizationParams(ctx, headBlock, headState)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"params.Spines":           params.Spines,
					"params.BaseSpine":        params.BaseSpine.Hex(),
					"checkpoint":              params.Checkpoint.Epoch,
					"params.Checkpoint.Spine": params.Checkpoint.Spine,
				}).Warn("dag finalization: get finalization params failed")
				return errors.Wrap(err, "dag finalization: get finalization params failed")
			}

			finRes, err := s.cfg.ExecutionEngineCaller.ExecutionDagFinalize(ctx, params)
			baseSpine := params.BaseSpine
			finalizing := params.Spines
			lfSpine := finRes.LFSpine
			fSeq := append(gwatCommon.HashArray{*baseSpine}, finalizing...)
			if err != nil || lfSpine == nil {

				log.WithError(err).WithFields(logrus.Fields{
					"params.Spines":           params.Spines,
					"params.BaseSpine":        params.BaseSpine.Hex(),
					"checkpoint":              params.Checkpoint.Epoch,
					"params.Checkpoint.Spine": params.Checkpoint.Spine,
					"lfSpine":                 lfSpine.Hex(),
				}).Warn("dag finalization: finalization failed")
				return errors.Wrap(err, "dag finalization: gwat finalization failed")
			}
			for _, h := range fSeq {
				finalizedSeq = append(finalizedSeq, h)
				if h == *lfSpine {
					break
				}
			}

			if len(finalizedSeq) == 0 {
				err = errors.New("lf spine is invalid")
				log.WithError(err).WithFields(logrus.Fields{
					"finalizationSeq": fSeq,
					"lfSpine":         lfSpine.Hex(),
					"isValid":         fSeq.Has(*lfSpine),
				}).Warn("dag finalization: finalization failed")
				return errors.Wrap(err, "dag finalization: gwat finalization failed")
			}

			log.WithFields(logrus.Fields{
				"finalized":        finalizing,
				"baseSpine":        baseSpine.Hex(),
				"lfSpine":          lfSpine.Hex(),
				"finalizationSeq":  fSeq,
				"finalizedSeq":     finalizedSeq,
				"isFullyFinalized": fSeq.IsEqualTo(finalizedSeq),
			}).Info("dag finalization: finalization success")
		}
	}

	//update checkpoint of FinalizedSpines cache
	cpFin := headState.FinalizedCheckpoint()
	cpState, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(cpFin.Root))
	if err != nil {
		log.WithError(errors.Wrapf(err, "Cache finalized spines: could not get checkpoint state for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot())).Error("dag finalization")
		return errors.Wrapf(err, "Cache finalized spines: could not get checkpoint state for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot())
	}
	if cpState == nil || cpState.IsNil() {
		log.WithError(errors.Wrapf(err, "Cache finalized spines: checkpoint's state not found for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot())).Error("dag finalization")
		return errors.Wrapf(err, "Cache finalized spines: checkpoint's state not found for epoch=%d root=%x", cpFin.Epoch, cpFin.GetRoot())
	}
	cpFinSeq := gwatCommon.HashArrayFromBytes(cpState.Eth1Data().Finalization)
	s.SetFinalizedSpinesCheckpoint(cpFinSeq[len(cpFinSeq)-1])

	//update FinalizedSpines cache
	s.AddFinalizedSpines(finalizedSeq)

	return nil
}

// collectFinalizationParams collects params to call gwat finalization api.
func (s *Service) collectFinalizationParams(
	ctx context.Context,
	headBlock block.SignedBeaconBlock,
	headState state.BeaconState,
) (*gwatTypes.FinalizationParams, error) {
	if headState == nil || headState.IsNil() {
		return nil, errors.New("Collect finalization params: nil head state received")
	}
	var (
		baseSpine       gwatCommon.Hash
		finalizationSeq gwatCommon.HashArray
		err             error
	)

	// get request gwat checkpoint
	checkpoint := headState.FinalizedCheckpoint()
	cpFinalized, err := s.getRequestGwatCheckpoint(ctx, headState)
	if err != nil {
		return nil, err
	}

	cpSlot, err := slots.EpochStart(checkpoint.Epoch)
	if err != nil {
		return nil, err
	}
	finalizedSpines := s.GetFinalizedSpines()
	var currRoot [32]byte
	currState := headState
	currBlock := headBlock
	for {
		currFinalization := gwatCommon.HashArrayFromBytes(currState.Eth1Data().Finalization)
		intersect := finalizedSpines.SequenceIntersection(currFinalization)
		if len(intersect) == 0 {
			//finalizationSeq = append(finalizationSeq, currFinalization...)
			finalizationSeq = append(currFinalization, finalizationSeq...)
		} else {
			baseSpine = intersect[len(intersect)-1]
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
			//update FinalizedSpines cache
			s.SetFinalizedSpinesHead(baseSpine)
			break
		}
		//set next block root as current
		currRoot = bytesutil.ToBytes32(currBlock.Block().ParentRoot())
		if currRoot == params.BeaconConfig().ZeroHash {
			return &gwatTypes.FinalizationParams{
				Spines:        finalizationSeq,
				BaseSpine:     &baseSpine,
				Checkpoint:    cpFinalized,
				ValidatorSync: nil,
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
			log.WithError(err).Error("Collect finalization params")
			return nil, err
		}
		if currState == nil || currState.IsNil() {
			err = errors.Errorf("retrieved nil parent state for root=%x", currRoot)
			log.WithError(err).Error("Collect finalization params")
			return nil, err
		}
		// if reach finalized checkpoint slot
		if currBlock.Block().Slot() < cpSlot {
			err = errors.New("Collect finalization params: failed")
			log.WithError(err).Error("Collect finalization params")
			return nil, err
		}
	}
	return &gwatTypes.FinalizationParams{
		Spines:        finalizationSeq,
		BaseSpine:     &baseSpine,
		Checkpoint:    cpFinalized,
		ValidatorSync: nil,
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
	// create
	cpRoot := bytesutil.ToBytes32(checkpoint.Root)
	cpState, err := s.cfg.StateGen.StateByRoot(ctx, cpRoot)
	if err != nil {
		return cpFinalized, err
	}

	finalization := gwatCommon.HashArrayFromBytes(cpState.Eth1Data().Finalization)
	cpFinalized = &gwatTypes.Checkpoint{
		Epoch: uint64(checkpoint.Epoch),
		Root:  gwatCommon.BytesToHash(checkpoint.Root),
		Spine: finalization[len(finalization)-1], // use last spine
	}
	s.CacheGwatCheckpoint(cpFinalized)

	return cpFinalized, nil
}
