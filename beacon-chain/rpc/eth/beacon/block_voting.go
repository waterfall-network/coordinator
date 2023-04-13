package beacon

import (
	"context"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/rpc/eth/helpers"
	ethpbv "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/eth/v1"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListBlockVotings retrieves the block votings for the given state.
// If the epoch is not passed in, then the block votings for the epoch of the state will be obtained.
func (bs *Server) ListBlockVotings(ctx context.Context, req *ethpbv.StateBlockVotingsRequest) (*ethpbv.StateBlockVotingsResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.ListBlockVotings")
	defer span.End()
	st, err := bs.StateFetcher.State(ctx, req.StateId)
	if err != nil {
		return nil, helpers.PrepareStateFetchGRPCError(err)
	}

	isOptimistic, err := helpers.IsOptimistic(ctx, st, bs.HeadFetcher)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not check if slot's block is optimistic: %v", err)
	}

	blockVotings := st.BlockVoting()
	data := make([]*ethpbv.BlockVoting, len(blockVotings))
	for _, blockVoting := range blockVotings {
		attestations := make([]*ethpbv.Attestation, len(blockVoting.Attestations))
		for _, att := range blockVoting.Attestations {
			attestations = append(attestations, &ethpbv.Attestation{
				AggregationBits: att.AggregationBits,
				Data: &ethpbv.AttestationData{
					Slot:            att.Data.Slot,
					Index:           att.Data.CommitteeIndex,
					BeaconBlockRoot: att.Data.BeaconBlockRoot,
					Source: &ethpbv.Checkpoint{
						Epoch: att.Data.Source.Epoch,
						Root:  att.Data.Source.Root,
					},
					Target: &ethpbv.Checkpoint{
						Epoch: att.Data.Target.Epoch,
						Root:  att.Data.Target.Root,
					},
				},
				Signature: att.Signature,
			})
		}

		data = append(data, &ethpbv.BlockVoting{
			Root:         blockVoting.Root,
			Slot:         blockVoting.Slot,
			Candidates:   blockVoting.Candidates,
			Attestations: attestations,
		})
	}

	return &ethpbv.StateBlockVotingsResponse{
		Data:                data,
		ExecutionOptimistic: isOptimistic,
	}, nil
}
