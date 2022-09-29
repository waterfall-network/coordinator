package node

import (
	ethpbservice "github.com/waterfall-foundation/coordinator/proto/eth/service"
)

var _ ethpbservice.BeaconNodeServer = (*Server)(nil)
