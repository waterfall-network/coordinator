package validator

import (
	"context"
	"fmt"

	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

func (vs *Server) getBellatrixBeaconBlock(ctx context.Context, req *ethpb.BlockRequest) (*ethpb.BeaconBlockBellatrix, error) {
	return nil, fmt.Errorf("deprecated")
}
