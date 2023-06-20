package validator

import (
	"context"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed"
	opfeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/operation"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProposeExit proposes an exit for a validator.
func (vs *Server) ProposeExit(ctx context.Context, req *ethpb.VoluntaryExit) (*ethpb.ProposeExitResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "nil request")
	}
	s, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}

	// Confirm the validator is eligible to exit with the parameters provided.
	val, err := s.ValidatorAtIndexReadOnly(req.ValidatorIndex)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validator index exceeds validator set length")
	}

	if err := blocks.VerifyExitData(val, s.Slot(), req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	vs.OperationNotifier.OperationFeed().Send(&feed.Event{
		Type: opfeed.ExitReceived,
		Data: &opfeed.ExitReceivedData{
			Exit: req,
		},
	})

	vs.ExitPool.InsertVoluntaryExit(ctx, s, req)

	r, err := req.HashTreeRoot()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get tree hash of exit: %v", err)
	}

	return &ethpb.ProposeExitResponse{
		ExitRoot: r[:],
	}, vs.P2P.Broadcast(ctx, req)
}
