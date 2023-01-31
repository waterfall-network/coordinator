// Package slasher defines a gRPC server implementation of a slasher service
// which allows for checking if attestations or blocks are slashable.
package slasher

import (
	slasherservice "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/slasher"
)

// Server defines a server implementation of the gRPC slasher service.
type Server struct {
	SlashingChecker slasherservice.SlashingChecker
}
