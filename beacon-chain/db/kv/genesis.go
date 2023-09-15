package kv

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	dbIface "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db/iface"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	statev1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
)

// SaveGenesisData bootstraps the beaconDB with a given genesis state.
func (s *Store) SaveGenesisData(ctx context.Context, genesisState state.BeaconState) error {
	stateRoot, err := genesisState.HashTreeRoot(ctx)
	if err != nil {
		return err
	}
	genesisBlk := blocks.NewGenesisBlock(stateRoot[:])
	genesisBlkRoot, err := genesisBlk.Block.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "could not get genesis block root")
	}
	wsb, err := wrapper.WrappedSignedBeaconBlock(genesisBlk)
	if err != nil {
		return errors.Wrap(err, "could not wrap genesis block")
	}
	if err := s.SaveBlock(ctx, wsb); err != nil {
		return errors.Wrap(err, "could not save genesis block")
	}
	if err := s.SaveState(ctx, genesisState, genesisBlkRoot); err != nil {
		return errors.Wrap(err, "could not save genesis state")
	}
	if err := s.SaveStateSummary(ctx, &ethpb.StateSummary{
		Slot: 0,
		Root: genesisBlkRoot[:],
	}); err != nil {
		return err
	}

	if err := s.SaveHeadBlockRoot(ctx, genesisBlkRoot); err != nil {
		return errors.Wrap(err, "could not save head block root")
	}
	if err := s.SaveGenesisBlockRoot(ctx, genesisBlkRoot); err != nil {
		return errors.Wrap(err, "could not save genesis block root")
	}
	return nil
}

// LoadGenesis loads a genesis state from a ssz-serialized byte slice, if no genesis exists already.
func (s *Store) LoadGenesis(ctx context.Context, sb []byte) error {
	st := &ethpb.BeaconState{}
	if err := st.UnmarshalSSZ(sb); err != nil {
		return err
	}
	gs, err := statev1.InitializeFromProtoUnsafe(st)
	if err != nil {
		return err
	}
	existing, err := s.GenesisState(ctx)
	if err != nil {
		return err
	}
	// If some different genesis state existed already, return an error. The same genesis state is
	// considered a no-op.
	if existing != nil && !existing.IsNil() {
		a, err := existing.HashTreeRoot(ctx)
		if err != nil {
			return err
		}
		b, err := gs.HashTreeRoot(ctx)
		if err != nil {
			return err
		}
		if a == b {
			return nil
		}
		log.WithError(dbIface.ErrExistingGenesisState).WithFields(logrus.Fields{
			"exist": fmt.Sprintf("%#x", a),
			"calc":  fmt.Sprintf("%#x", b),
		}).Error("Load genesis failed")
		return dbIface.ErrExistingGenesisState
	}

	if !bytes.Equal(gs.Fork().CurrentVersion, params.BeaconConfig().GenesisForkVersion) {
		return fmt.Errorf("loaded genesis fork version (%#x) does not match config genesis "+
			"fork version (%#x)", gs.Fork().CurrentVersion, params.BeaconConfig().GenesisForkVersion)
	}
	return s.SaveGenesisData(ctx, gs)
}

// EnsureEmbeddedGenesis checks that a genesis block has been generated when an embedded genesis
// state is used. If a genesis block does not exist, but a genesis state does, then we should call
// SaveGenesisData on the existing genesis state.
func (s *Store) EnsureEmbeddedGenesis(ctx context.Context) error {
	gb, err := s.GenesisBlock(ctx)
	if err != nil {
		return err
	}
	if gb != nil && !gb.IsNil() {
		return nil
	}
	gs, err := s.GenesisState(ctx)
	if err != nil {
		return err
	}
	if gs != nil && !gs.IsNil() {
		return s.SaveGenesisData(ctx, gs)
	}
	return nil
}
