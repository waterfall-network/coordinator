package attestations

import (
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/attestations/kv"
)

var _ Pool = (*kv.AttCaches)(nil)
