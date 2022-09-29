package attestations

import (
	"github.com/waterfall-foundation/coordinator/beacon-chain/operations/attestations/kv"
)

var _ Pool = (*kv.AttCaches)(nil)
