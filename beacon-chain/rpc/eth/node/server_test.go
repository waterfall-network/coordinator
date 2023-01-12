package node

import (
	ethpbservice "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/eth/service"
)

var _ ethpbservice.BeaconNodeServer = (*Server)(nil)
