package beacon

import ethpbservice "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/eth/service"

var _ ethpbservice.BeaconChainServer = (*Server)(nil)
