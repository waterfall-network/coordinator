package beacon

import (
	"context"

	ethpbv "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/eth/v1"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/rpc/eth/helpers"
)

// GetEth1Data retrieves the eth1 data for the given state.
// If the epoch is not passed in, then the eth1 data for the epoch of the state will be obtained.
func (bs *Server) GetEth1Data(ctx context.Context, req *ethpbv.StateEth1DataRequest) (*ethpbv.StateEth1DataResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.ListEth1Data")
	defer span.End()
	st, err := bs.StateFetcher.State(ctx, req.StateId)
	if err != nil {
		return nil, helpers.PrepareStateFetchGRPCError(err)
	}

	isOptimistic, err := helpers.IsOptimistic(ctx, st, bs.HeadFetcher)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not check if slot's block is optimistic: %v", err)
	}

	eth1Data := st.Eth1Data()

	return &ethpbv.StateEth1DataResponse{
		Data: &ethpbv.Eth1Data{
			DepositRoot:  eth1Data.DepositRoot,
			DepositCount: eth1Data.DepositCount,
			BlockHash:    eth1Data.BlockHash,
			Candidates:   eth1Data.Candidates,
			// TODO: uncomment when Finalization will be available, if it will be ever (comes from sync-validators branch)
			//Finalization: eth1Data.Finalization,
		},
		ExecutionOptimistic: isOptimistic,
	}, nil
}
