package sync

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"google.golang.org/protobuf/proto"
)

func (s *Service) voluntaryExitSubscriber(ctx context.Context, msg proto.Message) error {
	ve, ok := msg.(*ethpb.SignedVoluntaryExit)
	if !ok {
		return fmt.Errorf("wrong type, expected: *ethpb.SignedVoluntaryExit got: %T", msg)
	}

	if ve.Exit == nil {
		return errors.New("exit can't be nil")
	}
	s.setExitIndexSeen(ve.Exit.ValidatorIndex)

	headState, err := s.cfg.chain.HeadState(ctx)
	if err != nil {
		return err
	}
	s.cfg.exitPool.InsertVoluntaryExit(ctx, headState, ve)
	return nil
}

func (s *Service) committeeIndexBeaconPrevoteSubscriber(_ context.Context, msg proto.Message) error {
	prevote, ok := msg.(*ethpb.PreVote)
	if !ok {
		return fmt.Errorf("wrong type, expected: *ethpb.PreVote got: %T", msg)
	}

	if prevote.Data == nil {
		return errors.New("prevote data is nil")
	}

	exists, err := s.cfg.prevotePool.HasAggregatedPrevote(prevote)
	if err != nil {
		return errors.Wrap(err, "Could not determine if prevote pool has this prevote")
	}
	if exists {
		return nil
	}

	return s.cfg.prevotePool.SaveUnaggregatedPrevote(prevote)
}

func (s *Service) attesterSlashingSubscriber(ctx context.Context, msg proto.Message) error {
	aSlashing, ok := msg.(*ethpb.AttesterSlashing)
	if !ok {
		return fmt.Errorf("wrong type, expected: *ethpb.AttesterSlashing got: %T", msg)
	}
	// Do some nil checks to prevent easy DoS'ing of this handler.
	aSlashing1IsNil := aSlashing == nil || aSlashing.Attestation_1 == nil || aSlashing.Attestation_1.AttestingIndices == nil
	aSlashing2IsNil := aSlashing == nil || aSlashing.Attestation_2 == nil || aSlashing.Attestation_2.AttestingIndices == nil
	if !aSlashing1IsNil && !aSlashing2IsNil {
		headState, err := s.cfg.chain.HeadState(ctx)
		if err != nil {
			return err
		}
		if err := s.cfg.slashingPool.InsertAttesterSlashing(ctx, headState, aSlashing); err != nil {
			return errors.Wrap(err, "could not insert attester slashing into pool")
		}
		s.setAttesterSlashingIndicesSeen(aSlashing.Attestation_1.AttestingIndices, aSlashing.Attestation_2.AttestingIndices)
	}
	return nil
}

func (s *Service) proposerSlashingSubscriber(ctx context.Context, msg proto.Message) error {
	pSlashing, ok := msg.(*ethpb.ProposerSlashing)
	if !ok {
		return fmt.Errorf("wrong type, expected: *ethpb.ProposerSlashing got: %T", msg)
	}
	// Do some nil checks to prevent easy DoS'ing of this handler.
	header1IsNil := pSlashing == nil || pSlashing.Header_1 == nil || pSlashing.Header_1.Header == nil
	header2IsNil := pSlashing == nil || pSlashing.Header_2 == nil || pSlashing.Header_2.Header == nil
	if !header1IsNil && !header2IsNil {
		headState, err := s.cfg.chain.HeadState(ctx)
		if err != nil {
			return err
		}
		if err := s.cfg.slashingPool.InsertProposerSlashing(ctx, headState, pSlashing); err != nil {
			return errors.Wrap(err, "could not insert proposer slashing into pool")
		}
		s.setProposerSlashingIndexSeen(pSlashing.Header_1.Header.ProposerIndex)
	}
	return nil
}
