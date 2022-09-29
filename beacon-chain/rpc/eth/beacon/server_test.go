package beacon

import ethpbservice "github.com/waterfall-foundation/coordinator/proto/eth/service"

var _ ethpbservice.BeaconChainServer = (*Server)(nil)
